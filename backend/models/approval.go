package models

import (
	"time"
)

// ApprovalLog 审批流转日志模型
type ApprovalLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ReservationID uint      `gorm:"not null;index" json:"reservation_id"`
	ApproverID    uint      `gorm:"not null" json:"approver_id"`
	ApproverName  string    `gorm:"size:50" json:"approver_name"`
	ApproverRole  Role      `gorm:"size:20" json:"approver_role"` // "advisor" 或 "equipment_admin"
	Action        string    `gorm:"size:20;not null" json:"action"` // "approve" 或 "reject"
	Comment       string    `gorm:"size:255" json:"comment"`
	CreatedAt     time.Time `json:"created_at"`
}
