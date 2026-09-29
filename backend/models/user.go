package models

import (
	"time"
)

type Role string

const (
	RoleStudent        Role = "student"         // 学生
	RoleAdvisor        Role = "advisor"         // 导师/课题组长
	RoleEquipmentAdmin Role = "equipment_admin" // 设备管理员
	RoleSystemAdmin    Role = "system_admin"    // 系统管理员
)

// User 用户模型
type User struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Username    string    `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Password    string    `gorm:"size:100;not null" json:"-"`
	RealName    string    `gorm:"size:50;not null" json:"real_name"`
	Email       string    `gorm:"size:100" json:"email"`
	Phone       string    `gorm:"size:20" json:"phone"`
	Role        Role      `gorm:"size:20;not null;default:'student'" json:"role"`
	GroupID     *uint     `json:"group_id"` // 所属课题组ID
	GroupName   string    `gorm:"size:100" json:"group_name"`
	CreditScore int       `gorm:"default:100" json:"credit_score"` // 科研信誉分（默认100）
	IsFrozen    bool      `gorm:"default:false" json:"is_frozen"`  // 是否因信誉过低被冻结
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
