package services

import (
	"errors"
	"time"

	"lab-reservation-system/database"
	"lab-reservation-system/models"

	"gorm.io/gorm"
)

type CreditEngine struct{}

var DefaultCreditEngine = &CreditEngine{}

// CheckUserReservationEligibility 检查用户信誉是否具备预约资格
func (ce *CreditEngine) CheckUserReservationEligibility(userID uint, safetyLevel int) error {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	if user.IsFrozen || user.CreditScore < 60 {
		return errors.New("由于您的科研信誉积分低于60分，已被系统冻结预约权限，请联系实验室管理员核验处理")
	}

	if safetyLevel >= 2 && user.CreditScore < 80 {
		return errors.New("高精密/特种设备（安全等级≥2）要求科研信誉积分不低于80分，您当前信誉分不足")
	}

	return nil
}

// AdjustCreditScore 调整用户信誉积分并生成履约审计流水
func (ce *CreditEngine) AdjustCreditScore(tx *gorm.DB, userID uint, reservationID uint, delta int, reason string, recordType string) (int, error) {
	if tx == nil {
		tx = database.DB
	}

	var user models.User
	if err := tx.First(&user, userID).Error; err != nil {
		return 0, err
	}

	newScore := user.CreditScore + delta
	if newScore > 100 {
		newScore = 100
	}
	if newScore < 0 {
		newScore = 0
	}

	isFrozen := newScore < 60

	if err := tx.Model(&user).Updates(map[string]interface{}{
		"credit_score": newScore,
		"is_frozen":    isFrozen,
	}).Error; err != nil {
		return 0, err
	}

	// 记录信誉变动明细
	creditRecord := models.CreditRecord{
		UserID:        userID,
		UserName:      user.RealName,
		ReservationID: reservationID,
		DeltaScore:    delta,
		ScoreAfter:    newScore,
		Reason:        reason,
		Type:          recordType,
		CreatedAt:     time.Now(),
	}
	if err := tx.Create(&creditRecord).Error; err != nil {
		return 0, err
	}

	return newScore, nil
}
