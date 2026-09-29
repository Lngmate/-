package controllers

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"lab-reservation-system/database"
	"lab-reservation-system/models"
	"lab-reservation-system/services"

	"github.com/gin-gonic/gin"
)

var equipmentLocks sync.Map

func getEquipmentLock(equipmentID uint) *sync.Mutex {
	val, _ := equipmentLocks.LoadOrStore(equipmentID, &sync.Mutex{})
	return val.(*sync.Mutex)
}

type ReservationController struct{}

// CreateReservationRequest 创建预约请求载荷
type CreateReservationRequest struct {
	EquipmentID    uint      `json:"equipment_id" binding:"required"`
	StartTime      time.Time `json:"start_time" binding:"required"`
	EndTime        time.Time `json:"end_time" binding:"required"`
	ExperimentName string    `json:"experiment_name" binding:"required"`
	Purpose        string    `json:"purpose"`
	TargetTemp     int       `json:"target_temp"`
	VacuumNeeded   bool      `json:"vacuum_needed"`
}

// CheckConflictAndBatch 预先检测冲突并评估伴热拼单优惠（前端即时交互）
func (rc *ReservationController) CheckConflictAndBatch(c *gin.Context) {
	var req CreateReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数格式错误"})
		return
	}

	conflict, conflicting, err := services.DefaultConflictEngine.CheckConflict(nil, req.EquipmentID, req.StartTime, req.EndTime, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	if conflict {
		c.JSON(http.StatusOK, gin.H{
			"code":        409,
			"conflict":    true,
			"message":     fmt.Sprintf("所选时间段与已有预约 [%s] 冲突！", conflicting.ExperimentName),
			"conflicting": conflicting,
		})
		return
	}

	// 检测伴热拼单节能机会（创新点）
	batchOpp := services.DefaultConflictEngine.DetectBatchOpportunity(req.EquipmentID, req.StartTime, req.EndTime, req.TargetTemp, req.VacuumNeeded)

	c.JSON(http.StatusOK, gin.H{
		"code":     200,
		"conflict": false,
		"batching": batchOpp,
		"message":  "所选时间段空闲，可预约！",
	})
}

// Create 提交设备预约申请
func (rc *ReservationController) Create(c *gin.Context) {
	var req CreateReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请填写完整的预约信息"})
		return
	}

	// 高并发排期互斥锁：针对同一台仪器进行毫秒级加锁，彻底杜绝并发超约与重复排期
	eqLock := getEquipmentLock(req.EquipmentID)
	eqLock.Lock()
	defer eqLock.Unlock()

	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uint)

	// 1. 获取设备信息
	var equipment models.Equipment
	if err := database.DB.First(&equipment, req.EquipmentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "预约的设备不存在"})
		return
	}

	// 2. 检查用户信誉是否满足预约要求（信誉引擎）
	if err := services.DefaultCreditEngine.CheckUserReservationEligibility(userID, equipment.SafetyLevel); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": err.Error()})
		return
	}

	// 3. 校验时间冲突（防重叠数学引擎）
	conflict, conflicting, err := services.DefaultConflictEngine.CheckConflict(nil, req.EquipmentID, req.StartTime, req.EndTime, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	if conflict {
		c.JSON(http.StatusConflict, gin.H{
			"code":    409,
			"message": fmt.Sprintf("预约时间段冲突！已被 [%s] 抢占", conflicting.ExperimentName),
		})
		return
	}

	// 4. 获取用户信息与所属课题组
	var user models.User
	database.DB.First(&user, userID)
	if user.GroupID == nil || *user.GroupID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "您尚未绑定科研课题组，无法进行经费结算与预约，请先联系管理员分配"})
		return
	}

	// 5. 计算预约机时与费用
	durationHours := req.EndTime.Sub(req.StartTime).Hours()
	discountRate := 1.0
	isBatched := false

	// 检测伴热拼单优惠
	batchOpp := services.DefaultConflictEngine.DetectBatchOpportunity(req.EquipmentID, req.StartTime, req.EndTime, req.TargetTemp, req.VacuumNeeded)
	if batchOpp.HasOpportunity {
		discountRate = batchOpp.DiscountRate
		isBatched = true
	}

	totalCost := durationHours * equipment.HourlyRate * discountRate

	// 6. 生成 6 位随机核销签到码
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	checkinCode := fmt.Sprintf("%06d", r.Intn(1000000))

	// 7. 初始状态判定（基于设备审批策略）
	var initialStatus models.ReservationStatus
	switch equipment.ApprovalType {
	case models.ApprovalNone:
		initialStatus = models.ReservationApproved // 免审直通
	case models.ApprovalDual:
		initialStatus = models.ReservationPendingAdvisor // 双审：先导师
	default:
		initialStatus = models.ReservationPendingAdmin // 单审：管理员
	}

	res := models.Reservation{
		EquipmentID:    equipment.ID,
		EquipmentName:  equipment.Name,
		EquipmentModel: equipment.Model,
		UserID:         user.ID,
		UserName:       user.RealName,
		UserPhone:      user.Phone,
		GroupID:        *user.GroupID,
		GroupName:      user.GroupName,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		DurationHours:  durationHours,
		ExperimentName: req.ExperimentName,
		Purpose:        req.Purpose,
		TargetTemp:     req.TargetTemp,
		VacuumNeeded:   req.VacuumNeeded,
		IsBatched:      isBatched,
		DiscountRate:   discountRate,
		TotalCost:      totalCost,
		Status:         initialStatus,
		CheckinCode:    checkinCode,
	}

	if err := database.DB.Create(&res).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存预约工单失败"})
		return
	}

	// 触发站内消息与红点通知
	if initialStatus == models.ReservationPendingAdvisor {
		// 通知导师
		services.DefaultNotificationService.NotifyGroupAdvisors(
			*user.GroupID,
			"【待审批提醒】学生提交新预约",
			fmt.Sprintf("学生【%s】提交了设备【%s】的预约申请，实验课题：%s，时段：%s 至 %s，请及时审批！",
				user.RealName, equipment.Name, req.ExperimentName,
				req.StartTime.Format("2006-01-02 15:04"), req.EndTime.Format("15:04")),
			models.NotifReservationPending,
			res.ID,
		)
		// 通知学生本人
		services.DefaultNotificationService.SendNotification(
			user.ID,
			"【预约提交成功】等待导师审批",
			fmt.Sprintf("您已成功提交设备【%s】的预约申请，已推送到导师待办列表进行经费与资质审核。", equipment.Name),
			models.NotifReservationPending,
			res.ID,
		)
	} else if initialStatus == models.ReservationPendingAdmin {
		// 通知管理员
		services.DefaultNotificationService.NotifyAdmins(
			"【待审批提醒】新预约待排期审核",
			fmt.Sprintf("学生【%s】申请预约了设备【%s】，时段：%s 至 %s，请前往审批工作台审核！",
				user.RealName, equipment.Name,
				req.StartTime.Format("2006-01-02 15:04"), req.EndTime.Format("15:04")),
			models.NotifReservationPending,
			res.ID,
		)
		// 通知学生本人
		services.DefaultNotificationService.SendNotification(
			user.ID,
			"【预约提交成功】等待管理员审核",
			fmt.Sprintf("您已成功提交设备【%s】的预约申请，已推送到设备管理员进行排期审核。", equipment.Name),
			models.NotifReservationPending,
			res.ID,
		)
	} else if initialStatus == models.ReservationApproved {
		// 免审直通生效，通知学生本人
		services.DefaultNotificationService.SendNotification(
			user.ID,
			"【预约成功】免审直通已生效",
			fmt.Sprintf("恭喜！您预约的设备【%s】免审直通已生效！现场签到核销码为【%s】，时段：%s 至 %s，请准时赴约！",
				equipment.Name, res.CheckinCode,
				req.StartTime.Format("2006-01-02 15:04"), req.EndTime.Format("15:04")),
			models.NotifReservationApproved,
			res.ID,
		)
		// 抄送管理员
		services.DefaultNotificationService.NotifyAdmins(
			"【新预约通知】免审直通设备已预约",
			fmt.Sprintf("学生【%s】预约了免审设备【%s】，时段：%s 至 %s",
				user.RealName, equipment.Name,
				req.StartTime.Format("2006-01-02 15:04"), req.EndTime.Format("15:04")),
			models.NotifSystemNotice,
			res.ID,
		)
	}

	msg := "预约申请提交成功！"
	if initialStatus == models.ReservationApproved {
		msg = "该设备支持免审直通，预约已自动生效！签到核销码已生成。"
	} else if initialStatus == models.ReservationPendingAdvisor {
		msg = "预约已提交，已推送到导师待办列表进行经费与资质审核。"
	} else {
		msg = "预约已提交，已推送到设备管理员待办列表进行排期审核。"
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": msg,
		"data":    res,
	})
}

// List 获取预约列表
func (rc *ReservationController) List(c *gin.Context) {
	scope := c.Query("scope") // "my" or "all"
	equipmentID := c.Query("equipment_id")
	status := c.Query("status")
	keyword := c.Query("keyword")

	query := database.DB.Model(&models.Reservation{})

	if scope == "my" {
		userID, _ := c.Get("user_id")
		query = query.Where("user_id = ?", userID)
	}

	if equipmentID != "" {
		query = query.Where("equipment_id = ?", equipmentID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		pattern := "%" + strings.TrimSpace(keyword) + "%"
		query = query.Where("experiment_name LIKE ? OR equipment_name LIKE ? OR user_name LIKE ?", pattern, pattern, pattern)
	}

	var reservations []models.Reservation
	query.Order("start_time DESC").Find(&reservations)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": reservations,
	})
}

// Cancel 用户主动取消预约
func (rc *ReservationController) Cancel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效预约ID"})
		return
	}

	var res models.Reservation
	if err := database.DB.First(&res, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "预约工单不存在"})
		return
	}

	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("role")
	role := userRole.(models.Role)

	// 权限检查：只有本人或管理员可取消
	if res.UserID != userID.(uint) && role != models.RoleSystemAdmin && role != models.RoleEquipmentAdmin {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您无权取消该预约"})
		return
	}

	// 状态检查：只有待审核或未开始的已审核预约可取消
	if res.Status != models.ReservationPendingAdvisor &&
		res.Status != models.ReservationPendingAdmin &&
		res.Status != models.ReservationApproved {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该预约处于进行中或已完结状态，无法取消"})
		return
	}

	type CancelReq struct {
		Reason string `json:"reason"`
	}
	var req CancelReq
	c.ShouldBindJSON(&req)

	reason := "用户主动取消"
	if req.Reason != "" {
		reason = req.Reason
	}

	res.Status = models.ReservationCancelled
	res.RejectReason = reason
	database.DB.Save(&res)

	// 提前取消释放时间槽奖励或记录
	services.DefaultCreditEngine.AdjustCreditScore(nil, res.UserID, res.ID, 1, "主动提前取消释放紧缺设备时间槽", "cancel_release")

	// 发送取消确认通知
	services.DefaultNotificationService.SendNotification(
		res.UserID,
		"【预约已取消】机位已成功释放",
		fmt.Sprintf("您已成功取消设备【%s】的预约，机位已重新释放回共享资源池，获得 +1 守信鼓励分。", res.EquipmentName),
		models.NotifReservationCancel,
		res.ID,
	)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "预约已成功取消并释放机位！主动释放时间槽获得 +1 信誉分鼓励。",
		"data":    res,
	})
}

// CalendarEvents 为 FullCalendar 日历提供格式化事件流（支持状态着色）
func (rc *ReservationController) CalendarEvents(c *gin.Context) {
	equipmentID := c.Query("equipment_id")

	query := database.DB.Model(&models.Reservation{})
	if equipmentID != "" && equipmentID != "0" {
		query = query.Where("equipment_id = ?", equipmentID)
	}

	var reservations []models.Reservation
	query.Find(&reservations)

	type CalendarEvent struct {
		ID              uint                   `json:"id"`
		Title           string                 `json:"title"`
		Start           string                 `json:"start"`
		End             string                 `json:"end"`
		BackgroundColor string                 `json:"backgroundColor"`
		BorderColor     string                 `json:"borderColor"`
		ExtendedProps   map[string]interface{} `json:"extendedProps"`
	}

	events := make([]CalendarEvent, 0)
	for _, r := range reservations {
		bgColor := "#909399"
		switch r.Status {
		case models.ReservationActive:
			bgColor = "#409EFF" // 运行中-科技蓝
		case models.ReservationApproved:
			bgColor = "#67C23A" // 已审批通过-成功绿
		case models.ReservationPendingAdmin, models.ReservationPendingAdvisor:
			bgColor = "#E6A23C" // 审批中-警示黄
		case models.ReservationCompleted:
			bgColor = "#909399" // 已完成-灰色
		case models.ReservationRejected, models.ReservationNoShow:
			bgColor = "#F56C6C" // 驳回/爽约-红色
		case models.ReservationCancelled:
			bgColor = "#C0C4CC" // 已取消-淡灰
		}

		title := fmt.Sprintf("[%s] %s (%s)", r.EquipmentName, r.ExperimentName, r.UserName)

		events = append(events, CalendarEvent{
			ID:              r.ID,
			Title:           title,
			Start:           r.StartTime.Format(time.RFC3339),
			End:             r.EndTime.Format(time.RFC3339),
			BackgroundColor: bgColor,
			BorderColor:     bgColor,
			ExtendedProps: map[string]interface{}{
				"reservation_id": r.ID,
				"equipment_name": r.EquipmentName,
				"user_name":      r.UserName,
				"status":         r.Status,
				"is_batched":     r.IsBatched,
				"target_temp":    r.TargetTemp,
				"total_cost":     r.TotalCost,
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": events,
	})
}

// Checkin 现场签到上机（核销）
func (rc *ReservationController) Checkin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	type CheckinReq struct {
		CheckinCode      string `json:"checkin_code" binding:"required"`
		ConditionCheckIn string `json:"condition_check_in"`
	}

	var req CheckinReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请输入6位现场核销码"})
		return
	}

	var res models.Reservation
	if err := database.DB.First(&res, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "预约工单不存在"})
		return
	}

	if res.Status != models.ReservationApproved {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该预约尚未通过审批或已核销"})
		return
	}

	if res.CheckinCode != req.CheckinCode {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "核销码不正确，请向现场核实"})
		return
	}

	now := time.Now()
	res.CheckinAt = &now
	res.ConditionCheckIn = req.ConditionCheckIn
	res.Status = models.ReservationActive
	database.DB.Save(&res)

	// 更新设备状态为使用中
	database.DB.Model(&models.Equipment{}).Where("id = ?", res.EquipmentID).Update("status", models.StatusInUse)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "现场签到成功，设备已上电，实验机时开始计时！",
		"data":    res,
	})
}

// Checkout 实验结束归还与验收
func (rc *ReservationController) Checkout(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	type CheckoutReq struct {
		ConditionCheckOut string `json:"condition_check_out"`
	}

	var req CheckoutReq
	c.ShouldBindJSON(&req)

	var res models.Reservation
	if err := database.DB.First(&res, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "预约工单不存在"})
		return
	}

	if res.Status != models.ReservationActive {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该工单非运行中状态"})
		return
	}

	now := time.Now()
	res.CheckoutAt = &now
	res.ConditionCheckOut = req.ConditionCheckOut
	res.Status = models.ReservationCompleted

	// 开启事务：保存工单、扣减课题组经费、增加信誉分、更新设备累计机时与状态
	tx := database.DB.Begin()
	if err := tx.Save(&res).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "归还失败"})
		return
	}

	// 扣减课题组经费与机时
	desc := fmt.Sprintf("实验结算: %s (设备: %s, 时长: %.1f小时)", res.ExperimentName, res.EquipmentName, res.DurationHours)
	if err := services.DefaultSettlementEngine.DeductGroupBalance(tx, res.GroupID, res.ID, res.TotalCost, res.DurationHours, desc); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "经费结算失败: " + err.Error()})
		return
	}

	// 履约奖励：按时归还增加 2 点信誉分
	services.DefaultCreditEngine.AdjustCreditScore(tx, res.UserID, res.ID, 2, "按时完成实验并提交归还验收评价", "on_time_return")

	// 更新设备：重置状态为空闲，累计机时增加
	var eq models.Equipment
	tx.First(&eq, res.EquipmentID)
	eq.Status = models.StatusAvailable
	eq.TotalHours += res.DurationHours
	// 每次使用微量消耗健康度（每10小时约-1分）
	healthLoss := int(res.DurationHours / 10.0)
	if healthLoss < 1 {
		healthLoss = 1
	}
	if eq.HealthScore > healthLoss {
		eq.HealthScore -= healthLoss
	}
	tx.Save(&eq)

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "实验已顺利归还并结算！信誉积分+2，感谢规范使用实验室设备。",
		"data":    res,
	})
}
