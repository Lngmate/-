package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	"gorm.io/gorm"
	"lab-reservation-system/models"
)

var ErrScheduleChanged = errors.New("排期已发生变化，请重新生成方案")

type TaskStepInput struct {
	EquipmentID     uint   `json:"equipment_id"`
	Name            string `json:"name"`
	DurationMinutes int    `json:"duration_minutes"`
	GapMinutes      int    `json:"gap_minutes"` // Minimum delay before this step; zero for the first.
}

type TaskRequest struct {
	Name          string          `json:"name"`
	Purpose       string          `json:"purpose"`
	EarliestStart time.Time       `json:"earliest_start"`
	Deadline      time.Time       `json:"deadline"`
	Steps         []TaskStepInput `json:"steps"`
	StartTimes    []time.Time     `json:"start_times"` // Required only when confirming a preview.
}

type PlannedStep struct {
	TaskStepInput
	EquipmentName string                `json:"equipment_name"`
	StartTime     time.Time             `json:"start_time"`
	EndTime       time.Time             `json:"end_time"`
	WaitMinutes   int                   `json:"wait_minutes"`
	TotalCost     float64               `json:"total_cost"`
	ApprovalType  models.ApprovalPolicy `json:"approval_type"`
}

type TaskPlan struct {
	Steps     []PlannedStep `json:"steps"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	TotalCost float64       `json:"total_cost"`
}

func (req TaskRequest) Validate(now time.Time) error {
	if strings.TrimSpace(req.Name) == "" || len([]rune(req.Name)) > 100 {
		return errors.New("请填写100字以内的实验任务名称")
	}
	if len(req.Steps) < 2 || len(req.Steps) > 3 {
		return errors.New("实验任务需要2至3道工序")
	}
	if req.EarliestStart.Before(now) || !req.Deadline.After(req.EarliestStart) || req.Deadline.Sub(req.EarliestStart) > 30*24*time.Hour {
		return errors.New("请选择未来30天以内的有效排期窗口")
	}
	if req.Deadline.After(now.Add(30 * 24 * time.Hour)) {
		return errors.New("最晚完成时间不能超过未来30天")
	}
	for i, step := range req.Steps {
		if step.EquipmentID == 0 || strings.TrimSpace(step.Name) == "" || len([]rune(step.Name)) > 80 {
			return fmt.Errorf("请完善第%d道工序的名称与设备", i+1)
		}
		if step.DurationMinutes < 30 || step.DurationMinutes > 24*60 || step.GapMinutes < 0 || step.GapMinutes > 7*24*60 {
			return errors.New("工序时长须为30至1440分钟，间隔须为0至10080分钟")
		}
		if i == 0 && step.GapMinutes != 0 {
			return errors.New("第一道工序无需前置间隔")
		}
	}
	return nil
}

func activeReservationStatuses() []models.ReservationStatus {
	return []models.ReservationStatus{models.ReservationPendingAdvisor, models.ReservationPendingAdmin, models.ReservationApproved, models.ReservationActive}
}

// EarliestSlot jumps to the end of each conflicting interval, allowing touching boundaries.
func EarliestSlot(start time.Time, duration time.Duration, busy []models.Reservation) time.Time {
	for _, reservation := range busy {
		if !reservation.EndTime.After(start) {
			continue
		}
		if !reservation.StartTime.Before(start.Add(duration)) {
			break
		}
		start = reservation.EndTime
	}
	return start
}

// PlanTask greedily finds each earliest finish. For a fixed sequence with minimum gaps,
// an earlier predecessor finish cannot reduce the feasible choices for its successor.
func PlanTask(db *gorm.DB, userID uint, req TaskRequest, exact bool, now time.Time) (TaskPlan, error) {
	plan := TaskPlan{Steps: []PlannedStep{}}
	if err := req.Validate(now); err != nil {
		return plan, err
	}
	if exact && len(req.StartTimes) != len(req.Steps) {
		return plan, errors.New("请先生成排期方案后再确认")
	}
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		return plan, err
	}
	if user.GroupID == nil || *user.GroupID == 0 {
		return plan, errors.New("请先绑定科研课题组再预约")
	}
	if user.IsFrozen || user.CreditScore < 60 {
		return plan, errors.New("当前账户暂无预约权限，请联系管理员")
	}
	ready := req.EarliestStart
	for i, step := range req.Steps {
		var eq models.Equipment
		if err := db.First(&eq, step.EquipmentID).Error; err != nil {
			return plan, fmt.Errorf("第%d道工序的设备不存在", i+1)
		}
		if eq.Status == models.StatusMaintenance || eq.Status == models.StatusFaulty || eq.Status == models.StatusScrapped {
			return plan, fmt.Errorf("设备【%s】暂停预约，请更换设备", eq.Name)
		}
		if eq.SafetyLevel >= 2 && user.CreditScore < 80 {
			return plan, fmt.Errorf("预约【%s】需要信誉分不低于80分", eq.Name)
		}
		ready = ready.Add(time.Duration(step.GapMinutes) * time.Minute)
		var busy []models.Reservation
		if err := db.Where("equipment_id = ? AND status IN ? AND end_time > ? AND start_time < ?", eq.ID, activeReservationStatuses(), ready, req.Deadline).Order("start_time ASC").Find(&busy).Error; err != nil {
			return plan, err
		}
		duration := time.Duration(step.DurationMinutes) * time.Minute
		start := EarliestSlot(ready, duration, busy)
		if exact {
			start = req.StartTimes[i]
			if start.Before(ready) {
				return plan, ErrScheduleChanged
			}
			for _, b := range busy {
				if b.StartTime.Before(start.Add(duration)) && b.EndTime.After(start) {
					return plan, ErrScheduleChanged
				}
			}
		}
		end := start.Add(duration)
		if end.After(req.Deadline) {
			return plan, fmt.Errorf("截止时间内无法安排第%d道工序【%s】，请扩大时间窗口或更换设备", i+1, step.Name)
		}
		cost := math.Round(float64(step.DurationMinutes)/60*eq.HourlyRate*100) / 100
		plan.Steps = append(plan.Steps, PlannedStep{TaskStepInput: step, EquipmentName: eq.Name, StartTime: start, EndTime: end, WaitMinutes: int(start.Sub(ready) / time.Minute), TotalCost: cost, ApprovalType: eq.ApprovalType})
		plan.TotalCost += cost
		ready = end
	}
	plan.StartTime = plan.Steps[0].StartTime
	plan.EndTime = ready
	plan.TotalCost = math.Round(plan.TotalCost*100) / 100
	return plan, nil
}

// SubmitTask revalidates the exact preview under the same locks as single bookings.
// Task, reservations and notification are committed together or rolled back together.
func SubmitTask(db *gorm.DB, userID uint, req TaskRequest) (models.ExperimentTask, error) {
	var task models.ExperimentTask
	ids := []uint{}
	for _, step := range req.Steps {
		ids = append(ids, step.EquipmentID)
	}
	unlock := LockEquipments(ids)
	defer unlock()
	err := db.Transaction(func(tx *gorm.DB) error {
		plan, err := PlanTask(tx, userID, req, true, time.Now())
		if err != nil {
			return err
		}
		var user models.User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}
		task = models.ExperimentTask{UserID: userID, Name: strings.TrimSpace(req.Name), Purpose: req.Purpose, EarliestStart: req.EarliestStart, Deadline: req.Deadline}
		if err := tx.Create(&task).Error; err != nil {
			return err
		}
		for i, step := range plan.Steps {
			var eq models.Equipment
			if err := tx.First(&eq, step.EquipmentID).Error; err != nil {
				return err
			}
			code, err := rand.Int(rand.Reader, big.NewInt(1000000))
			if err != nil {
				return err
			}
			status := models.ReservationPendingAdmin
			if eq.ApprovalType == models.ApprovalNone {
				status = models.ReservationApproved
			}
			if eq.ApprovalType == models.ApprovalDual {
				status = models.ReservationPendingAdvisor
			}
			reservation := models.Reservation{TaskID: &task.ID, TaskStep: i + 1, GapMinutes: step.GapMinutes, EquipmentID: eq.ID, EquipmentName: eq.Name, EquipmentModel: eq.Model, UserID: user.ID, UserName: user.RealName, UserPhone: user.Phone, GroupID: *user.GroupID, GroupName: user.GroupName, StartTime: step.StartTime, EndTime: step.EndTime, DurationHours: float64(step.DurationMinutes) / 60, ExperimentName: step.Name, Purpose: task.Name + "：" + req.Purpose, DiscountRate: 1, TotalCost: step.TotalCost, Status: status, CheckinCode: fmt.Sprintf("%06d", code.Int64())}
			if err := tx.Create(&reservation).Error; err != nil {
				return err
			}
			task.Reservations = append(task.Reservations, reservation)
		}
		notification := models.Notification{UserID: userID, Title: "实验任务预约已提交", Content: fmt.Sprintf("任务【%s】的%d道工序已提交，可前往实验任务查看各工序审批与履约进度。", task.Name, len(plan.Steps)), Type: models.NotifSystemNotice, RelatedID: task.Reservations[0].ID}
		return tx.Create(&notification).Error
	})
	return task, err
}
