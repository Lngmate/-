package services

import (
	"errors"
	"math"
	"time"

	"lab-reservation-system/database"
	"lab-reservation-system/models"

	"gorm.io/gorm"
)

// ConflictEngine 时间冲突检测引擎
type ConflictEngine struct{}

var DefaultConflictEngine = &ConflictEngine{}

// CheckConflict 检查目标设备在给定时间段是否存在时间重叠
// 算法依据：max(Start_New, Start_Exist) < min(End_New, End_Exist)
// 在SQL中等价于：start_time < newEnd AND end_time > newStart
func (e *ConflictEngine) CheckConflict(tx *gorm.DB, equipmentID uint, newStart, newEnd time.Time, excludeReservationID uint) (bool, *models.Reservation, error) {
	if tx == nil {
		tx = database.DB
	}

	// 基础合法性检验：开始时间必须早于结束时间，且必须在未来
	if !newStart.Before(newEnd) {
		return true, nil, errors.New("预约开始时间必须早于结束时间")
	}
	if newEnd.Sub(newStart) < 30*time.Minute {
		return true, nil, errors.New("单次预约时长不得低于30分钟")
	}

	// 查询有效预约状态（待审批、已通过、正在进行中）
	validStatuses := []models.ReservationStatus{
		models.ReservationPendingAdvisor,
		models.ReservationPendingAdmin,
		models.ReservationApproved,
		models.ReservationActive,
	}

	var conflicting models.Reservation
	query := tx.Where("equipment_id = ? AND status IN ? AND start_time < ? AND end_time > ?",
		equipmentID, validStatuses, newEnd, newStart)

	if excludeReservationID > 0 {
		query = query.Where("id <> ?", excludeReservationID)
	}

	err := query.First(&conflicting).Error
	if err == nil {
		// 存在冲突
		return true, &conflicting, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 无冲突
		return false, nil, nil
	}

	return true, nil, err
}

// BatchingOpportunity 伴热/同参拼单机会检测结果
type BatchingOpportunity struct {
	HasOpportunity     bool    `json:"has_opportunity"`
	AdjacentID         uint    `json:"adjacent_id"`
	AdjacentExperiment string  `json:"adjacent_experiment"`
	AdjacentTemp       int     `json:"adjacent_temp"`
	DiscountRate       float64 `json:"discount_rate"`
	Reason             string  `json:"reason"`
}

// DetectBatchOpportunity 检测前后相邻预约是否存在伴热/同工况拼单机会（创新点）
func (e *ConflictEngine) DetectBatchOpportunity(equipmentID uint, newStart, newEnd time.Time, targetTemp int, vacuumNeeded bool) BatchingOpportunity {
	var equipment models.Equipment
	if err := database.DB.First(&equipment, equipmentID).Error; err != nil || !equipment.SupportBatching {
		return BatchingOpportunity{HasOpportunity: false, DiscountRate: 1.0}
	}

	// 查找前后 60 分钟以内的邻近预约
	windowBefore := newStart.Add(-60 * time.Minute)
	windowAfter := newEnd.Add(60 * time.Minute)

	validStatuses := []models.ReservationStatus{
		models.ReservationApproved,
		models.ReservationActive,
		models.ReservationPendingAdmin,
	}

	var neighbors []models.Reservation
	database.DB.Where("equipment_id = ? AND status IN ? AND ((end_time >= ? AND end_time <= ?) OR (start_time >= ? AND start_time <= ?))",
		equipmentID, validStatuses, windowBefore, newStart, newEnd, windowAfter).
		Find(&neighbors)

	for _, n := range neighbors {
		// 判定工况相似度：如果最高温度>100℃且目标温度偏差在150℃以内，或都需要真空抽气
		tempDiff := math.Abs(float64(n.TargetTemp - targetTemp))
		if (targetTemp > 100 && tempDiff <= 150) || (vacuumNeeded && n.VacuumNeeded) {
			return BatchingOpportunity{
				HasOpportunity:     true,
				AdjacentID:         n.ID,
				AdjacentExperiment: n.ExperimentName,
				AdjacentTemp:       n.TargetTemp,
				DiscountRate:       0.85, // 享8.5折
				Reason:             "检测到紧密相邻且工况相近的高温/真空实验，触发伴热拼单节能优惠(8.5折)！",
			}
		}
	}

	return BatchingOpportunity{HasOpportunity: false, DiscountRate: 1.0}
}
