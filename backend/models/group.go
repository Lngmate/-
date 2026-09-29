package models

import (
	"time"
)

// ResearchGroup 课题组虚拟账户模型
type ResearchGroup struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Name           string    `gorm:"size:100;uniqueIndex;not null" json:"name"`
	LeaderID       uint      `json:"leader_id"`                                // 课题组长/导师用户ID
	LeaderName     string    `gorm:"size:50" json:"leader_name"`
	Department     string    `gorm:"size:100" json:"department"`               // 所属院系（如材料科学与工程学院）
	Balance        float64   `gorm:"default:10000" json:"balance"`             // 虚拟科研经费余额（元）
	TotalHoursUsed float64   `gorm:"default:0" json:"total_hours_used"`        // 累计消耗机时
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TransactionRecord 课题组虚拟账户消费与划转流水
type TransactionRecord struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	GroupID       uint      `gorm:"not null;index" json:"group_id"`
	GroupName     string    `gorm:"size:100" json:"group_name"`
	ReservationID uint      `json:"reservation_id"`
	Amount        float64   `gorm:"not null" json:"amount"`       // 金额变动（负数为扣费，正数为充值或退款）
	BalanceAfter  float64   `json:"balance_after"`                // 变动后余额
	Type          string    `gorm:"size:50" json:"type"`          // "reservation_fee", "recharge", "refund", "cross_group_settlement"
	Description   string    `gorm:"size:255" json:"description"`  // 摘要说明
	CreatedAt     time.Time `json:"created_at"`
}
