package models

import (
	"time"
)

// MaintenanceRecord 设备维护与检修日志模型
type MaintenanceRecord struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	EquipmentID   uint      `gorm:"not null;index" json:"equipment_id"`
	EquipmentName string    `gorm:"size:100" json:"equipment_name"`
	OperatorID    uint      `json:"operator_id"`
	OperatorName  string    `gorm:"size:50" json:"operator_name"`
	Type          string    `gorm:"size:50;not null" json:"type"` // "routine_inspection"(日常巡检), "repair"(故障维修), "calibration"(标定计量)
	Description   string    `gorm:"type:text" json:"description"`
	Cost          float64   `gorm:"default:0" json:"cost"`
	HealthScoreBefore int   `json:"health_score_before"`
	HealthScoreAfter  int   `json:"health_score_after"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	CreatedAt     time.Time `json:"created_at"`
}
