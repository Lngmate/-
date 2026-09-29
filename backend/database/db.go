package database

import (
	"fmt"
	"log"
	"time"

	"lab-reservation-system/config"
	"lab-reservation-system/models"
	"lab-reservation-system/utils"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化数据库连接并自动迁移表结构与初始种子数据
func InitDB(cfg *config.Config) *gorm.DB {
	var dialector gorm.Dialector

	if cfg.DBType == "mysql" {
		dialector = mysql.Open(cfg.DSN)
	} else {
		// 使用纯Go实现的CGO-free SQLite驱动，Windows环境无需任何额外编译工具
		dialector = sqlite.Open(cfg.DSN)
	}

	var err error
	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("无法连接到数据库 (%s): %v", cfg.DBType, err)
	}

	// 自动迁移表结构
	err = DB.AutoMigrate(
		&models.User{},
		&models.Equipment{},
		&models.Reservation{},
		&models.ResearchGroup{},
		&models.TransactionRecord{},
		&models.CreditRecord{},
		&models.ApprovalLog{},
		&models.MaintenanceRecord{},
		&models.Notification{},
	)
	if err != nil {
		log.Fatalf("数据库表结构自动迁移失败: %v", err)
	}

	// 填充基础初始化数据（演示与评测数据）
	seedData()

	return DB
}

// seedData 预置全套高校科研设备与演示数据集
func seedData() {
	var userCount int64
	DB.Model(&models.User{}).Count(&userCount)
	if userCount > 0 {
		return // 已经初始化过，跳过
	}

	log.Println("正在初始化高校实验室完整系统预置数据...")

	// 1. 初始化 5 大科研课题组
	g1 := models.ResearchGroup{Name: "先进纳米材料课题组", LeaderName: "李教授", Department: "材料科学与工程学院", Balance: 35000, TotalHoursUsed: 168.5}
	g2 := models.ResearchGroup{Name: "超导物理与量子计算组", LeaderName: "陈教授", Department: "物理学院", Balance: 28000, TotalHoursUsed: 112.0}
	g3 := models.ResearchGroup{Name: "生物催化与精准合成组", LeaderName: "王教授", Department: "化学与化工学院", Balance: 45000, TotalHoursUsed: 210.0}
	g4 := models.ResearchGroup{Name: "光电能源与微纳器件中心", LeaderName: "赵教授", Department: "光电信息工程学院", Balance: 36000, TotalHoursUsed: 145.5}
	g5 := models.ResearchGroup{Name: "智能仿生与机器人装备中心", LeaderName: "孙教授", Department: "机械与材料交叉学院", Balance: 50000, TotalHoursUsed: 190.0}
	DB.Create(&g1)
	DB.Create(&g2)
	DB.Create(&g3)
	DB.Create(&g4)
	DB.Create(&g5)

	// 2. 初始化预置用户（各角色覆盖，密码采用bcrypt加密）
	pwd, _ := utils.HashPassword("123")

	u1 := models.User{Username: "student", Password: pwd, RealName: "张耀辉(学生)", Role: models.RoleStudent, GroupID: &g1.ID, GroupName: g1.Name, CreditScore: 98, Email: "zhang@univ.edu.cn", Phone: "13800001111"}
	u2 := models.User{Username: "student2", Password: pwd, RealName: "周浩楠(学生)", Role: models.RoleStudent, GroupID: &g2.ID, GroupName: g2.Name, CreditScore: 88, Email: "zhou@univ.edu.cn", Phone: "13800002222"}
	u3 := models.User{Username: "student3", Password: pwd, RealName: "林思源(学生)", Role: models.RoleStudent, GroupID: &g4.ID, GroupName: g4.Name, CreditScore: 95, Email: "lin@univ.edu.cn", Phone: "13800003333"}
	u4 := models.User{Username: "advisor", Password: pwd, RealName: "李教授(导师/组长)", Role: models.RoleAdvisor, GroupID: &g1.ID, GroupName: g1.Name, CreditScore: 100, Email: "prof_li@univ.edu.cn", Phone: "13900003333"}
	u5 := models.User{Username: "advisor2", Password: pwd, RealName: "陈教授(导师/组长)", Role: models.RoleAdvisor, GroupID: &g2.ID, GroupName: g2.Name, CreditScore: 100, Email: "prof_chen@univ.edu.cn", Phone: "13900004444"}
	u6 := models.User{Username: "admin", Password: pwd, RealName: "王佳豪(设备管理员)", Role: models.RoleEquipmentAdmin, CreditScore: 100, Email: "admin_wang@univ.edu.cn", Phone: "13700004444"}
	u7 := models.User{Username: "sysadmin", Password: pwd, RealName: "李俊达(系统管理员)", Role: models.RoleSystemAdmin, CreditScore: 100, Email: "sysadmin@univ.edu.cn", Phone: "13600005555"}

	DB.Create(&u1)
	DB.Create(&u2)
	DB.Create(&u3)
	DB.Create(&u4)
	DB.Create(&u5)
	DB.Create(&u6)
	DB.Create(&u7)

	// 更新课题组组长ID
	DB.Model(&g1).Update("leader_id", u4.ID)
	DB.Model(&g2).Update("leader_id", u5.ID)

	// 3. 初始化 16 台大型高精密科研仪器（布局分布于 2D 孪生地图 A/B/C/D 四大功能区）
	equipments := []models.Equipment{
		// === A区：微观形貌与微观结构表征平台 ===
		{
			Name:             "场发射扫描电子显微镜(FE-SEM)",
			Model:            "JEOL JSM-7800F",
			Category:         "显微镜与微观表征",
			Location:         "理科实验大楼 301 室",
			LabRoom:          "Lab-301",
			MapX:             18,
			MapY:             22,
			SafetyLevel:      2,
			TrainingRequired: true,
			HourlyRate:       180.0,
			Status:           models.StatusAvailable,
			ApprovalType:     models.ApprovalDual,
			GroupID:          g1.ID,
			GroupName:        g1.Name,
			AdminID:          u6.ID,
			AdminName:        u6.RealName,
			HealthScore:      92,
			TotalHours:       340.5,
			MaxTemp:          0,
			VacuumRequired:   true,
			SupportBatching:  false,
			Specs:            "分辨率: 0.8nm @ 15kV, 1.2nm @ 1kV; 加速电压: 0.5 - 30kV; 配备牛津高灵敏能谱仪(EDS)",
		},
		{
			Name:             "高分辨透射电子显微镜(HR-TEM)",
			Model:            "FEI Talos F200X",
			Category:         "显微镜与微观表征",
			Location:         "理科实验大楼 302 室",
			LabRoom:          "Lab-302",
			MapX:             36,
			MapY:             22,
			SafetyLevel:      3,
			TrainingRequired: true,
			HourlyRate:       350.0,
			Status:           models.StatusAvailable,
			ApprovalType:     models.ApprovalDual,
			GroupID:          g1.ID,
			GroupName:        g1.Name,
			AdminID:          u6.ID,
			AdminName:        u6.RealName,
			HealthScore:      94,
			TotalHours:       280.0,
			MaxTemp:          0,
			VacuumRequired:   true,
			SupportBatching:  false,
			Specs:            "200kV场发射枪; 信息分辨率: 0.12nm; 搭载Super-X四探头无窗EDS系统及HAADF探头",
		},
		{
			Name:             "激光共聚焦显微镜(CLSM)",
			Model:            "Leica TCS SP8 STED",
			Category:         "显微镜与微观表征",
			Location:         "化学生物楼 201 室",
			LabRoom:          "Lab-201",
			MapX:             18,
			MapY:             40,
			SafetyLevel:      2,
			TrainingRequired: true,
			HourlyRate:       160.0,
			Status:           models.StatusInUse,
			ApprovalType:     models.ApprovalSingle,
			GroupID:          g3.ID,
			GroupName:        g3.Name,
			AdminID:          u6.ID,
			AdminName:        u6.RealName,
			HealthScore:      88,
			TotalHours:       310.0,
			MaxTemp:          37,
			VacuumRequired:   false,
			SupportBatching:  false,
			Specs:            "受激发射损耗超高分辨率成像; 扫描速度: 40幅/秒 @ 512x512; 配备活细胞恒温恒湿培养装置",
		},
		{
			Name:             "原子力显微镜(AFM)",
			Model:            "Bruker Dimension Icon",
			Category:         "显微镜与微观表征",
			Location:         "纳米新材料楼 101 室",
			LabRoom:          "Lab-101",
			MapX:             36,
			MapY:             40,
			SafetyLevel:      2,
			TrainingRequired: true,
			HourlyRate:       120.0,
			Status:           models.StatusAvailable,
			ApprovalType:     models.ApprovalSingle,
			GroupID:          g1.ID,
			GroupName:        g1.Name,
			AdminID:          u6.ID,
			AdminName:        u6.RealName,
			HealthScore:      96,
			TotalHours:       190.0,
			MaxTemp:          0,
			VacuumRequired:   false,
			SupportBatching:  false,
			Specs:            "PeakForce Tapping智能敲击模式; 闭环X-Y噪声 < 0.15nm; 支持高灵敏纳米电学/力学性能映射",
		},

		// === B区：晶体结构与物性表征平台 ===
		{
			Name:             "X射线多晶粉末衍射仪(XRD)",
			Model:            "Bruker D8 ADVANCE",
			Category:         "晶体结构与物性表征",
			Location:         "理科实验大楼 305 室",
			LabRoom:          "Lab-305",
			MapX:             64,
			MapY:             22,
			SafetyLevel:      2,
			TrainingRequired: false,
			HourlyRate:       120.0,
			Status:           models.StatusAvailable,
			ApprovalType:     models.ApprovalSingle,
			GroupID:          g2.ID,
			GroupName:        g2.Name,
			AdminID:          u6.ID,
			AdminName:        u6.RealName,
			HealthScore:      95,
			TotalHours:       215.0,
			MaxTemp:          0,
			VacuumRequired:   false,
			SupportBatching:  true, // 支持伴热/批量拼单
			Specs:            "Cu靶陶瓷X射线管; 测角仪半径: 280mm; 2θ角度范围: -110°至168°; LynxEye阵列探测器",
		},
		{
			Name:             "X射线光电子能谱仪(XPS)",
			Model:            "Thermo Fisher ESCALAB 250Xi",
			Category:         "晶体结构与物性表征",
			Location:         "精密仪器中心 501 室",
			LabRoom:          "Lab-501",
			MapX:             82,
			MapY:             22,
			SafetyLevel:      3,
			TrainingRequired: true,
			HourlyRate:       260.0,
			Status:           models.StatusAvailable,
			ApprovalType:     models.ApprovalDual,
			GroupID:          g1.ID,
			GroupName:        g1.Name,
			AdminID:          u6.ID,
			AdminName:        u6.RealName,
			HealthScore:      91,
			TotalHours:       240.0,
			MaxTemp:          0,
			VacuumRequired:   true,
			SupportBatching:  false,
			Specs:            "微聚焦单色化Al靶; 能量分辨率: 0.45eV; 配备氩团簇离子枪(MAGCIS)支持有机无损深度剖析",
		},
		{
			Name:             "综合物性测量系统(PPMS-9T)",
			Model:            "Quantum Design PPMS-9T",
			Category:         "晶体结构与物性表征",
			Location:         "理科实验大楼 402 室",
			LabRoom:          "Lab-402",
			MapX:             64,
			MapY:             40,
			SafetyLevel:      3,
			TrainingRequired: true,
			HourlyRate:       220.0,
			Status:           models.StatusAvailable,
			ApprovalType:     models.ApprovalDual,
			GroupID:          g2.ID,
			GroupName:        g2.Name,
			AdminID:          u6.ID,
			AdminName:        u6.RealName,
			HealthScore:      85,
			TotalHours:       320.0,
			MaxTemp:          400,
			VacuumRequired:   true,
			SupportBatching:  false,
			Specs:            "温区范围: 1.9K - 400K; 超导磁场: ±9T; 支持高精度交直流电输运、比热与磁光克尔测量",
		},
		{
			Name:             "稳态/瞬态荧光光谱仪",
			Model:            "Edinburgh FLS1000",
			Category:         "晶体结构与物性表征",
			Location:         "化学生物楼 205 室",
			LabRoom:          "Lab-205",
			MapX:             82,
			MapY:             40,
			SafetyLevel:      2,
			TrainingRequired: false,
			HourlyRate:       140.0,
			Status:           models.StatusAvailable,
			ApprovalType:     models.ApprovalSingle,
			GroupID:          g4.ID,
			GroupName:        g4.Name,
			AdminID:          u6.ID,
			AdminName:        u6.RealName,
			HealthScore:      97,
			TotalHours:       130.0,
			MaxTemp:          0,
			VacuumRequired:   false,
			SupportBatching:  true,
			Specs:            "测量波段: 200nm - 1700nm 近红外扩展; 荧光寿命测量范围: 50ps - 50s; 配备积分球绝对量子产率测试系统",
		},

		// === C区：高温合成与薄膜制备平台 ===
		{
			Name:             "超高温真空气氛管式炉",
			Model:            "NBD-T1700-60IT",
			Category:         "高温合成与薄膜制备",
			Location:         "先进材料馆 104 室",
			LabRoom:          "Lab-104",
			MapX:             18,
			MapY:             66,
			SafetyLevel:      3,
			TrainingRequired: true,
			HourlyRate:       50.0,
			Status:           models.StatusAvailable,
			ApprovalType:     models.ApprovalSingle,
			GroupID:          g1.ID,
			GroupName:        g1.Name,
			AdminID:          u6.ID,
			AdminName:        u6.RealName,
			HealthScore:      86,
			TotalHours:       420.0,
			MaxTemp:          1700,
			VacuumRequired:   true,
			SupportBatching:  true, // 伴热拼单8.5折
			Specs:            "最高工作温度: 1700℃; 温控精度: ±1℃; 双层壳体风冷结构; 配备高真空分子泵组",
		},
		{
			Name:             "高真空三靶磁控溅射镀膜仪",
			Model:            "JGP-450",
			Category:         "高温合成与薄膜制备",
			Location:         "先进材料馆 106 室",
			LabRoom:          "Lab-106",
			MapX:             36,
			MapY:             66,
			SafetyLevel:      2,
			TrainingRequired: true,
			HourlyRate:       150.0,
			Status:           models.StatusAvailable,
			ApprovalType:     models.ApprovalSingle,
			GroupID:          g4.ID,
			GroupName:        g4.Name,
			AdminID:          u6.ID,
			AdminName:        u6.RealName,
			HealthScore:      90,
			TotalHours:       260.0,
			MaxTemp:          800,
			VacuumRequired:   true,
			SupportBatching:  true,
			Specs:            "三阴极直流/射频共溅射; 极限真空: 5.0×10^-5 Pa; 基片加热温度可达800℃; 配备膜厚监控仪",
		},
		{
			Name:             "等离子体增强化学气相沉积系统(PECVD)",
			Model:            "Oxford PlasmaPro 80",
			Category:         "高温合成与薄膜制备",
			Location:         "先进材料馆 108 室",
			LabRoom:          "Lab-108",
			MapX:             18,
			MapY:             84,
			SafetyLevel:      3,
			TrainingRequired: true,
			HourlyRate:       180.0,
			Status:           models.StatusAvailable,
			ApprovalType:     models.ApprovalDual,
			GroupID:          g4.ID,
			GroupName:        g4.Name,
			AdminID:          u6.ID,
			AdminName:        u6.RealName,
			HealthScore:      76, // 预警健康度
			TotalHours:       380.0,
			MaxTemp:          400,
			VacuumRequired:   true,
			SupportBatching:  true,
			Specs:            "13.56MHz射频等离子体源; 沉积材料: SiO2, SiNx, a-Si; 样品尺寸兼容至8英寸晶圆",
		},
		{
			Name:             "真空冷冻干燥机",
			Model:            "Labconco FreeZone 6L",
			Category:         "高温合成与薄膜制备",
			Location:         "化学生物楼 209 室",
			LabRoom:          "Lab-209",
			MapX:             36,
			MapY:             84,
			SafetyLevel:      1,
			TrainingRequired: false,
			HourlyRate:       40.0,
			Status:           models.StatusAvailable,
			ApprovalType:     models.ApprovalNone, // 免审直通
			GroupID:          g3.ID,
			GroupName:        g3.Name,
			AdminID:          u6.ID,
			AdminName:        u6.RealName,
			HealthScore:      98,
			TotalHours:       150.0,
			MaxTemp:          45,
			VacuumRequired:   true,
			SupportBatching:  true,
			Specs:            "冷阱最低温度: -84℃; 冰凝结容量: 6L/24小时; 聚四氟乙烯防腐蚀涂层; 适合生物纳米材料无损脱水",
		},

		// === D区：生化分离与色谱质谱平台 ===
		{
			Name:             "超临界流体色谱-质谱联用仪(SFC-MS)",
			Model:            "Waters ACQUITY UPC2",
			Category:         "生化分离与质谱色谱",
			Location:         "化学生物楼 208 室",
			LabRoom:          "Lab-208",
			MapX:             64,
			MapY:             66,
			SafetyLevel:      2,
			TrainingRequired: true,
			HourlyRate:       200.0,
			Status:           models.StatusInUse,
			ApprovalType:     models.ApprovalDual,
			GroupID:          g3.ID,
			GroupName:        g3.Name,
			AdminID:          u6.ID,
			AdminName:        u6.RealName,
			HealthScore:      89,
			TotalHours:       180.0,
			MaxTemp:          0,
			VacuumRequired:   false,
			SupportBatching:  false,
			Specs:            "以超临界CO2为流动相; 进样精度极高; 配备Q-Da高灵敏质谱检测器; 适合手性药物快速拆分",
		},
		{
			Name:             "600MHz超导核磁共振波谱仪(NMR)",
			Model:            "Bruker AVANCE NEO 600",
			Category:         "生化分离与质谱色谱",
			Location:         "精密仪器中心 101 室",
			LabRoom:          "Lab-101B",
			MapX:             82,
			MapY:             66,
			SafetyLevel:      3,
			TrainingRequired: true,
			HourlyRate:       280.0,
			Status:           models.StatusAvailable,
			ApprovalType:     models.ApprovalDual,
			GroupID:          g3.ID,
			GroupName:        g3.Name,
			AdminID:          u6.ID,
			AdminName:        u6.RealName,
			HealthScore:      93,
			TotalHours:       310.0,
			MaxTemp:          0,
			VacuumRequired:   false,
			SupportBatching:  false,
			Specs:            "14.1特斯拉屏蔽超导磁体; 搭载5mm Prodigy液氮超低温冷探头; 支持1H/13C/19F/31P多核高分辨谱图采集",
		},
		{
			Name:             "高速冷冻离心机",
			Model:            "Beckman Coulter Avanti J-26S",
			Category:         "生化分离与质谱色谱",
			Location:         "化学生物楼 210 室",
			LabRoom:          "Lab-210",
			MapX:             64,
			MapY:             84,
			SafetyLevel:      1,
			TrainingRequired: false,
			HourlyRate:       30.0,
			Status:           models.StatusAvailable,
			ApprovalType:     models.ApprovalNone, // 免审直通
			GroupID:          g3.ID,
			GroupName:        g3.Name,
			AdminID:          u6.ID,
			AdminName:        u6.RealName,
			HealthScore:      98,
			TotalHours:       95.0,
			MaxTemp:          40,
			VacuumRequired:   true,
			SupportBatching:  true,
			Specs:            "最高转速: 26,000 rpm; 最大相对离心力: 81,800 x g; 温控范围: -10℃ 至 +40℃",
		},
		{
			Name:             "电感耦合等离子体质谱仪(ICP-MS)",
			Model:            "Agilent 7900",
			Category:         "生化分离与质谱色谱",
			Location:         "环境与化学实验楼 303 室",
			LabRoom:          "Lab-303",
			MapX:             82,
			MapY:             84,
			SafetyLevel:      2,
			TrainingRequired: true,
			HourlyRate:       240.0,
			Status:           models.StatusAvailable,
			ApprovalType:     models.ApprovalSingle,
			GroupID:          g5.ID,
			GroupName:        g5.Name,
			AdminID:          u6.ID,
			AdminName:        u6.RealName,
			HealthScore:      90,
			TotalHours:       220.0,
			MaxTemp:          0,
			VacuumRequired:   true,
			SupportBatching:  false,
			Specs:            "超高基体耐受技术(UHMI); 检出限达 ppt 级别; 10个数量级线性动态范围; 全自动调谐与碰撞反应池",
		},
	}

	for i := range equipments {
		DB.Create(&equipments[i])
	}

	// 4. 初始化多状态仿真预约工单
	now := time.Now()

	// 运行中 1 (SFC-MS)
	startActive1 := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, time.Local)
	endActive1 := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, time.Local)
	r1 := models.Reservation{
		EquipmentID:      equipments[12].ID,
		EquipmentName:    equipments[12].Name,
		EquipmentModel:   equipments[12].Model,
		UserID:           u2.ID,
		UserName:         u2.RealName,
		UserPhone:        u2.Phone,
		GroupID:          g2.ID,
		GroupName:        g2.Name,
		StartTime:        startActive1,
		EndTime:          endActive1,
		DurationHours:    3.0,
		ExperimentName:   "手性药物异构体高效纯化与质谱定性",
		Purpose:          "国家自然科学基金项目测试",
		Status:           models.ReservationActive,
		TotalCost:        600.0,
		CheckinCode:      "849201",
		CheckinAt:        &startActive1,
		ConditionCheckIn: "真空密封良好，柱压稳定，流动相充足",
	}
	DB.Create(&r1)

	// 运行中 2 (CLSM)
	startActive2 := time.Date(now.Year(), now.Month(), now.Day(), 10, 30, 0, 0, time.Local)
	endActive2 := time.Date(now.Year(), now.Month(), now.Day(), 13, 30, 0, 0, time.Local)
	r2 := models.Reservation{
		EquipmentID:      equipments[2].ID,
		EquipmentName:    equipments[2].Name,
		EquipmentModel:   equipments[2].Model,
		UserID:           u3.ID,
		UserName:         u3.RealName,
		UserPhone:        u3.Phone,
		GroupID:          g4.ID,
		GroupName:        g4.Name,
		StartTime:        startActive2,
		EndTime:          endActive2,
		DurationHours:    3.0,
		ExperimentName:   "仿生微纳马达活细胞内吞示踪实验",
		Purpose:          "重点研发计划关键指标验证",
		Status:           models.ReservationActive,
		TotalCost:        480.0,
		CheckinCode:      "512930",
		CheckinAt:        &startActive2,
		ConditionCheckIn: "激光器输出稳定，恒温箱温度控制在37℃",
	}
	DB.Create(&r2)

	// 当天下午预约（待管理员审批，伴热拼单8.5折）
	startPending := time.Date(now.Year(), now.Month(), now.Day(), 14, 0, 0, 0, time.Local)
	endPending := time.Date(now.Year(), now.Month(), now.Day(), 17, 0, 0, 0, time.Local)
	r3 := models.Reservation{
		EquipmentID:    equipments[8].ID,
		EquipmentName:  equipments[8].Name,
		EquipmentModel: equipments[8].Model,
		UserID:         u1.ID,
		UserName:       u1.RealName,
		UserPhone:      u1.Phone,
		GroupID:        g1.ID,
		GroupName:      g1.Name,
		StartTime:      startPending,
		EndTime:        endPending,
		DurationHours:  3.0,
		ExperimentName: "二维过渡金属硫族化物CVD高温气相沉积",
		Purpose:        "毕业课题样品制备",
		TargetTemp:     1100,
		VacuumNeeded:   true,
		IsBatched:      true, // 伴热拼单享受8.5折
		DiscountRate:   0.85,
		TotalCost:      127.5,
		Status:         models.ReservationPendingAdmin,
		CheckinCode:    "392817",
	}
	DB.Create(&r3)

	// 明天上午预约（已审批通过待签到，FE-SEM）
	tomorrow := now.AddDate(0, 0, 1)
	startApproved := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 9, 30, 0, 0, time.Local)
	endApproved := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 11, 30, 0, 0, time.Local)
	r4 := models.Reservation{
		EquipmentID:    equipments[0].ID,
		EquipmentName:  equipments[0].Name,
		EquipmentModel: equipments[0].Model,
		UserID:         u1.ID,
		UserName:       u1.RealName,
		UserPhone:      u1.Phone,
		GroupID:        g1.ID,
		GroupName:      g1.Name,
		StartTime:      startApproved,
		EndTime:        endApproved,
		DurationHours:  2.0,
		ExperimentName: "石墨烯基催化剂微观形貌与EDS能谱定性",
		Purpose:        "高水平论文返修补做电镜形貌表征",
		TotalCost:      360.0,
		Status:         models.ReservationApproved,
		CheckinCode:    "628190",
	}
	DB.Create(&r4)

	// 明天下午预约（待导师审核，HR-TEM双审）
	startDual := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 14, 0, 0, 0, time.Local)
	endDual := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 16, 30, 0, 0, time.Local)
	r5 := models.Reservation{
		EquipmentID:    equipments[1].ID,
		EquipmentName:  equipments[1].Name,
		EquipmentModel: equipments[1].Model,
		UserID:         u1.ID,
		UserName:       u1.RealName,
		UserPhone:      u1.Phone,
		GroupID:        g1.ID,
		GroupName:      g1.Name,
		StartTime:      startDual,
		EndTime:        endDual,
		DurationHours:  2.5,
		ExperimentName: "高熵合金原子相界与位错微观晶格观察",
		Purpose:        "国家杰青基金项目子课题测试",
		TotalCost:      875.0,
		Status:         models.ReservationPendingAdvisor,
		CheckinCode:    "419208",
	}
	DB.Create(&r5)

	// 历史已完成工单（XRD）
	yesterday := now.AddDate(0, 0, -1)
	startPast := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 14, 0, 0, 0, time.Local)
	endPast := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 16, 0, 0, 0, time.Local)
	r6 := models.Reservation{
		EquipmentID:       equipments[4].ID,
		EquipmentName:     equipments[4].Name,
		EquipmentModel:    equipments[4].Model,
		UserID:            u1.ID,
		UserName:          u1.RealName,
		UserPhone:         u1.Phone,
		GroupID:           g1.ID,
		GroupName:         g1.Name,
		StartTime:         startPast,
		EndTime:           endPast,
		DurationHours:     2.0,
		ExperimentName:    "钙钛矿太阳能电池吸光层物相晶化分析",
		Purpose:           "验证退火温度对衍射峰强度影响",
		Status:            models.ReservationCompleted,
		TotalCost:         240.0,
		CheckinCode:       "771239",
		CheckinAt:         &startPast,
		CheckoutAt:        &endPast,
		ConditionCheckIn:  "测角仪初始归零正常，X射线快门无异常",
		ConditionCheckOut: "样品无损取出，清理样品托盘，工作区无残留",
	}
	DB.Create(&r6)

	// 5. 初始化信誉积分流水
	c1 := models.CreditRecord{
		UserID:     u1.ID,
		UserName:   u1.RealName,
		DeltaScore: 2,
		ScoreAfter: 98,
		Reason:     "按时履约完成实验并如实记录设备完好归还状态",
		Type:       "on_time_return",
		CreatedAt:  now.AddDate(0, 0, -2),
	}
	c2 := models.CreditRecord{
		UserID:     u2.ID,
		UserName:   u2.RealName,
		DeltaScore: -12,
		ScoreAfter: 88,
		Reason:     "预约FE-SEM超时25分钟未签到，触发迟到扣分预警",
		Type:       "late",
		CreatedAt:  now.AddDate(0, 0, -5),
	}
	DB.Create(&c1)
	DB.Create(&c2)

	// 6. 初始化课题组虚拟结算流水
	t1 := models.TransactionRecord{
		GroupID:       g1.ID,
		GroupName:     g1.Name,
		ReservationID: r3.ID,
		Amount:        -127.5,
		BalanceAfter:  34872.5,
		Type:          "reservation_fee",
		Description:   "预约管式炉机时费(伴热拼单8.5折优惠)",
		CreatedAt:     now,
	}
	t2 := models.TransactionRecord{
		GroupID:       g1.ID,
		GroupName:     g1.Name,
		Amount:        10000.0,
		BalanceAfter:  35000.0,
		Type:          "recharge",
		Description:   "先进材料国家重点实验室年度科研机时补贴",
		CreatedAt:     now.AddDate(0, 0, -10),
	}
	DB.Create(&t1)
	DB.Create(&t2)

	// 7. 初始化设备维保审计记录
	m1 := models.MaintenanceRecord{
		EquipmentID:       equipments[8].ID,
		EquipmentName:     equipments[8].Name,
		OperatorID:        u6.ID,
		OperatorName:      u6.RealName,
		Type:              "routine",
		Description:       "管式炉刚玉炉膛高温洁净化烧结与热电偶校准",
		Cost:              320.0,
		HealthScoreBefore: 75,
		HealthScoreAfter:  100,
		StartTime:         now.AddDate(0, 0, -7),
		EndTime:           now.AddDate(0, 0, -7).Add(3 * time.Hour),
		CreatedAt:         now.AddDate(0, 0, -7),
	}
	DB.Create(&m1)

	// 8. 初始化历史审批流日志
	appr1 := models.ApprovalLog{
		ReservationID: r4.ID,
		ApproverID:    u6.ID,
		ApproverName:  u6.RealName,
		ApproverRole:  models.RoleEquipmentAdmin,
		Action:        "approve",
		Comment:       "排期无冲突，已核准学生安全培训记录，准予使用FE-SEM",
		CreatedAt:     now.AddDate(0, 0, -1),
	}
	appr2 := models.ApprovalLog{
		ReservationID: r6.ID,
		ApproverID:    u6.ID,
		ApproverName:  u6.RealName,
		ApproverRole:  models.RoleEquipmentAdmin,
		Action:        "approve",
		Comment:       "XRD常温测试排期合规，同意预约",
		CreatedAt:     now.AddDate(0, 0, -2),
	}
	DB.Create(&appr1)
	DB.Create(&appr2)

	log.Println("系统预置数据初始化成功！")
	fmt.Println("==================================================")
	fmt.Println("高校实验室设备预约与共享系统预置账号：")
	fmt.Println("1. 学生:         student      密码: 123 (信誉分: 98)")
	fmt.Println("2. 学生(候补):   student2     密码: 123 (信誉分: 88)")
	fmt.Println("3. 学生(光电):   student3     密码: 123 (信誉分: 95)")
	fmt.Println("4. 导师/组长:    advisor      密码: 123 (先进纳米材料组)")
	fmt.Println("5. 导师/组长2:   advisor2     密码: 123 (超导物理量子组)")
	fmt.Println("6. 设备管理员:   admin        密码: 123 (审批与机位管理)")
	fmt.Println("7. 系统管理员:   sysadmin     密码: 123 (系统全局运维)")
	fmt.Println("==================================================")
}
