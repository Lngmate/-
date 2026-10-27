package models

import (
	"time"
)

type ReservationStatus string

const (
	ReservationPendingAdvisor ReservationStatus = "pending_advisor" // 待导师审批
	ReservationPendingAdmin   ReservationStatus = "pending_admin"   // 待管理员审批
	ReservationApproved       ReservationStatus = "approved"        // 审批通过（待履约）
	ReservationActive         ReservationStatus = "active"          // 实验运行中（已签到）
	ReservationCompleted      ReservationStatus = "completed"       // 实验结束（已归还核销）
	ReservationRejected       ReservationStatus = "rejected"        // 审批驳回
	ReservationCancelled      ReservationStatus = "cancelled"       // 用户取消
	ReservationNoShow         ReservationStatus = "no_show"         // 爽约（超期未签到）
)

// Reservation 设备时间槽预约工单模型
type Reservation struct {
	TaskID         *uint  `gorm:"index" json:"task_id"`
	TaskStep       int    `json:"task_step"`
	GapMinutes     int    `json:"gap_minutes"`
	ID             uint   `gorm:"primaryKey" json:"id"`
	EquipmentID    uint   `gorm:"not null;index" json:"equipment_id"`
	EquipmentName  string `gorm:"size:100" json:"equipment_name"`
	EquipmentModel string `gorm:"size:100" json:"equipment_model"`
	UserID         uint   `gorm:"not null;index" json:"user_id"`
	UserName       string `gorm:"size:50" json:"user_name"`
	UserPhone      string `gorm:"size:20" json:"user_phone"`
	GroupID        uint   `gorm:"not null;index" json:"group_id"`
	GroupName      string `gorm:"size:100" json:"group_name"`

	StartTime     time.Time `gorm:"not null;index" json:"start_time"`
	EndTime       time.Time `gorm:"not null;index" json:"end_time"`
	DurationHours float64   `json:"duration_hours"`

	ExperimentName string `gorm:"size:100;not null" json:"experiment_name"` // 实验课题名称
	Purpose        string `gorm:"type:text" json:"purpose"`                 // 实验用途简述

	// 创新点：伴热拼单参数与节能优惠
	TargetTemp   int     `json:"target_temp"`                      // 工况温度要求
	VacuumNeeded bool    `json:"vacuum_needed"`                    // 是否需要抽真空
	IsBatched    bool    `gorm:"default:false" json:"is_batched"`  // 是否为伴热/同参拼单
	DiscountRate float64 `gorm:"default:1.0" json:"discount_rate"` // 折扣率(如0.85)
	TotalCost    float64 `gorm:"default:0" json:"total_cost"`      // 预计/结算扣减经费

	// 状态与审批
	Status       ReservationStatus `gorm:"size:30;not null;default:'pending_admin'" json:"status"`
	RejectReason string            `gorm:"size:255" json:"reject_reason"`

	// 现场履约与核销
	CheckinCode       string     `gorm:"size:10" json:"checkin_code"`         // 6位动态核销码
	CheckinAt         *time.Time `json:"checkin_at"`                          // 实际签到上机时间
	CheckoutAt        *time.Time `json:"checkout_at"`                         // 实际归还下机时间
	ConditionCheckIn  string     `gorm:"size:255" json:"condition_check_in"`  // 初始状态确认
	ConditionCheckOut string     `gorm:"size:255" json:"condition_check_out"` // 归还验收状况

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
