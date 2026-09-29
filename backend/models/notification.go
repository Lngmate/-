package models

import (
	"time"
)

// NotificationType 消息通知类型
type NotificationType string

const (
	NotifReservationPending  NotificationType = "reservation_pending"  // 待审批提醒（推给导师/管理员）
	NotifReservationApproved NotificationType = "reservation_approved" // 审批通过提醒（推给学生）
	NotifReservationRejected NotificationType = "reservation_rejected" // 审批驳回提醒（推给学生）
	NotifReservationCancel   NotificationType = "reservation_cancel"   // 预约取消提醒
	NotifSystemNotice        NotificationType = "system_notice"        // 系统公告/通知
)

// Notification 站内消息与红点通知模型
type Notification struct {
	ID        uint             `gorm:"primaryKey" json:"id"`
	UserID    uint             `gorm:"not null;index" json:"user_id"` // 接收人ID
	Title     string           `gorm:"size:100;not null" json:"title"`
	Content   string           `gorm:"size:500;not null" json:"content"`
	Type      NotificationType `gorm:"size:30;not null" json:"type"`
	RelatedID uint             `gorm:"default:0" json:"related_id"` // 关联的预约工单ID或实体ID
	IsRead    bool             `gorm:"default:false;index" json:"is_read"`
	CreatedAt time.Time        `json:"created_at"`
}
