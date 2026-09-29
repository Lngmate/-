package models

import (
	"time"
)

// CreditRecord 科研履约信用奖惩流水模型
type CreditRecord struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"not null;index" json:"user_id"`
	UserName      string    `gorm:"size:50" json:"user_name"`
	ReservationID uint      `json:"reservation_id"`
	DeltaScore    int       `gorm:"not null" json:"delta_score"` // 变动分值（如 -10, +2）
	ScoreAfter    int       `json:"score_after"`                 // 变动后积分
	Reason        string    `gorm:"size:255;not null" json:"reason"`
	Type          string    `gorm:"size:50" json:"type"`         // "no_show", "late", "damage", "on_time_return", "admin_adjustment"
	CreatedAt     time.Time `json:"created_at"`
}
