package services

import (
	"errors"
	"time"

	"lab-reservation-system/database"
	"lab-reservation-system/models"

	"gorm.io/gorm"
)

type SettlementEngine struct{}

var DefaultSettlementEngine = &SettlementEngine{}

// DeductGroupBalance 扣减课题组虚拟经费与累计机时
func (se *SettlementEngine) DeductGroupBalance(tx *gorm.DB, groupID uint, reservationID uint, amount float64, durationHours float64, desc string) error {
	if tx == nil {
		tx = database.DB
	}

	var group models.ResearchGroup
	if err := tx.First(&group, groupID).Error; err != nil {
		return errors.New("课题组账户不存在")
	}

	if group.Balance < amount {
		return errors.New("课题组虚拟账户余额不足，请联系导师/课题组长充值或分配机时")
	}

	newBalance := group.Balance - amount
	newHours := group.TotalHoursUsed + durationHours

	if err := tx.Model(&group).Updates(map[string]interface{}{
		"balance":          newBalance,
		"total_hours_used": newHours,
	}).Error; err != nil {
		return err
	}

	// 记录流水
	record := models.TransactionRecord{
		GroupID:       groupID,
		GroupName:     group.Name,
		ReservationID: reservationID,
		Amount:        -amount,
		BalanceAfter:  newBalance,
		Type:          "reservation_fee",
		Description:   desc,
		CreatedAt:     time.Now(),
	}

	return tx.Create(&record).Error
}
