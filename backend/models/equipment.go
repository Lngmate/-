package models

import (
	"time"
)

type EquipmentStatus string

const (
	StatusAvailable   EquipmentStatus = "available"   // 空闲就绪
	StatusReserved    EquipmentStatus = "reserved"    // 已被预约
	StatusInUse       EquipmentStatus = "in_use"      // 实验运行中
	StatusMaintenance EquipmentStatus = "maintenance" // 维保检修
	StatusFaulty      EquipmentStatus = "faulty"      // 故障停用
	StatusScrapped    EquipmentStatus = "scrapped"    // 报废处置
)

type ApprovalPolicy string

const (
	ApprovalNone   ApprovalPolicy = "none"   // 免审直通
	ApprovalSingle ApprovalPolicy = "single" // 单审（设备管理员）
	ApprovalDual   ApprovalPolicy = "dual"   // 双审（导师审经费资质+管理员审机时）
)

// Equipment 实验设备全生命周期模型
type Equipment struct {
	ID               uint            `gorm:"primaryKey" json:"id"`
	Name             string          `gorm:"size:100;not null" json:"name"`
	Model            string          `gorm:"size:100" json:"model"`            // 规格型号
	Category         string          `gorm:"size:50" json:"category"`          // 分类（如光谱仪、显微镜、高温炉）
	Location         string          `gorm:"size:100;not null" json:"location"`// 物理位置（如材料馆302）
	LabRoom          string          `gorm:"size:50" json:"lab_room"`          // 实验室编号（如Lab-302）
	MapX             int             `json:"map_x"`                            // 2D数字孪生地图X坐标(百分比0-100)
	MapY             int             `json:"map_y"`                            // 2D数字孪生地图Y坐标(百分比0-100)
	SafetyLevel      int             `gorm:"default:1" json:"safety_level"`    // 安全等级：1(基础), 2(精密), 3(危险/特种)
	TrainingRequired bool            `gorm:"default:false" json:"training_required"` // 是否需要上机培训认证
	HourlyRate       float64         `gorm:"default:0" json:"hourly_rate"`     // 每小时虚拟机时单价（元/时）
	Status           EquipmentStatus `gorm:"size:20;default:'available'" json:"status"`
	ApprovalType     ApprovalPolicy  `gorm:"size:20;default:'single'" json:"approval_type"` // 审批策略
	
	// 归属与管理
	GroupID          uint            `json:"group_id"`                         // 出资/归属课题组ID
	GroupName        string          `gorm:"size:100" json:"group_name"`
	AdminID          uint            `json:"admin_id"`                         // 负责管理员ID
	AdminName        string          `gorm:"size:50" json:"admin_name"`
	
	// 创新点参数：健康度预测与伴热工况
	HealthScore      int             `gorm:"default:100" json:"health_score"`  // 设备健康度得分(0-100)
	TotalHours       float64         `gorm:"default:0" json:"total_hours"`     // 累计运行机时(小时)
	MaxTemp          int             `json:"max_temp"`                         // 最高工况温度(℃)
	VacuumRequired   bool            `gorm:"default:false" json:"vacuum_required"` // 是否需要真空抽气
	SupportBatching  bool            `gorm:"default:true" json:"support_batching"` // 是否支持伴热/同参拼单优惠
	
	Specs            string          `gorm:"type:text" json:"specs"`           // 技术指标与参数说明
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}
