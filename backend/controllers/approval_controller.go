package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"lab-reservation-system/database"
	"lab-reservation-system/models"
	"lab-reservation-system/services"

	"github.com/gin-gonic/gin"
)

type ApprovalController struct{}

// PendingList 获取待当前角色审批的申请列表
func (ac *ApprovalController) PendingList(c *gin.Context) {
	roleVal, _ := c.Get("role")
	role := roleVal.(models.Role)
	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uint)

	var list []models.Reservation

	if role == models.RoleAdvisor {
		// 导师审批自己课题组内的申请
		var user models.User
		database.DB.First(&user, userID)
		if user.GroupID != nil {
			database.DB.Where("group_id = ? AND status = ?", *user.GroupID, models.ReservationPendingAdvisor).
				Order("created_at DESC").Find(&list)
		}
	} else if role == models.RoleEquipmentAdmin || role == models.RoleSystemAdmin {
		// 设备管理员与系统管理员审批待管理员审核的申请
		database.DB.Where("status = ?", models.ReservationPendingAdmin).
			Order("created_at DESC").Find(&list)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": list,
	})
}

// Approve 审批通过
func (ac *ApprovalController) Approve(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	roleVal, _ := c.Get("role")
	role := roleVal.(models.Role)
	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uint)
	realNameVal, _ := c.Get("real_name")
	realName := realNameVal.(string)

	type ActionReq struct {
		Comment string `json:"comment"`
	}
	var req ActionReq
	c.ShouldBindJSON(&req)

	var res models.Reservation
	if err := database.DB.First(&res, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "预约申请不存在"})
		return
	}

	// 记录审批流日志
	approvalLog := models.ApprovalLog{
		ReservationID: res.ID,
		ApproverID:    userID,
		ApproverName:  realName,
		ApproverRole:  role,
		Action:        "approve",
		Comment:       req.Comment,
		CreatedAt:     time.Now(),
	}
	database.DB.Create(&approvalLog)

	// 根据当前状态推进工作流
	if res.Status == models.ReservationPendingAdvisor {
		// 导师审批通过 -> 转入管理员审批
		res.Status = models.ReservationPendingAdmin
		database.DB.Save(&res)

		// 通知学生
		services.DefaultNotificationService.SendNotification(
			res.UserID,
			"【审批进展】导师审批已通过",
			fmt.Sprintf("您的导师【%s】已审核通过您对设备【%s】的预约申请，工单已流转至设备管理员处进行排期审核。", realName, res.EquipmentName),
			models.NotifReservationApproved,
			res.ID,
		)
		// 通知管理员
		services.DefaultNotificationService.NotifyAdmins(
			"【待排期审核】导师已审核通过，待管理员终审",
			fmt.Sprintf("学生【%s】关于设备【%s】的预约已获导师批准，请管理员前往审批工作台确认排期！", res.UserName, res.EquipmentName),
			models.NotifReservationPending,
			res.ID,
		)

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "导师审批已通过，已推送到设备管理员进行最终排期审核",
			"data":    res,
		})
		return
	} else if res.Status == models.ReservationPendingAdmin {
		// 管理员审批通过 -> 最终生效，等待签到履约
		res.Status = models.ReservationApproved
		database.DB.Save(&res)

		// 通知学生：包含现场签到码和时间
		services.DefaultNotificationService.SendNotification(
			res.UserID,
			"【🎉预约审批通过】您的预约已正式生效！",
			fmt.Sprintf("恭喜！设备管理员【%s】已批准您的设备【%s】预约申请！现场签到核销码为【%s】，预约时段：%s 至 %s，请准时到场上机！",
				realName, res.EquipmentName, res.CheckinCode,
				res.StartTime.Format("2006-01-02 15:04"), res.EndTime.Format("15:04")),
			models.NotifReservationApproved,
			res.ID,
		)

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": fmt.Sprintf("设备管理员审批通过，预约已正式生效！现场签到码为: %s", res.CheckinCode),
			"data":    res,
		})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "当前预约状态无需审批"})
}

// Reject 审批驳回
func (ac *ApprovalController) Reject(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	roleVal, _ := c.Get("role")
	role := roleVal.(models.Role)
	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uint)
	realNameVal, _ := c.Get("real_name")
	realName := realNameVal.(string)

	type ActionReq struct {
		Comment string `json:"comment" binding:"required"`
	}
	var req ActionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请填写驳回原因"})
		return
	}

	var res models.Reservation
	if err := database.DB.First(&res, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "预约申请不存在"})
		return
	}

	res.Status = models.ReservationRejected
	res.RejectReason = req.Comment
	database.DB.Save(&res)

	// 记录审批日志
	approvalLog := models.ApprovalLog{
		ReservationID: res.ID,
		ApproverID:    userID,
		ApproverName:  realName,
		ApproverRole:  role,
		Action:        "reject",
		Comment:       req.Comment,
		CreatedAt:     time.Now(),
	}
	database.DB.Create(&approvalLog)

	// 通知申请学生
	services.DefaultNotificationService.SendNotification(
		res.UserID,
		"【⚠️预约驳回通知】您的预约申请被驳回",
		fmt.Sprintf("很遗憾，您关于设备【%s】的预约申请被【%s】驳回。驳回原因：%s。您可在预约列表中调整时间后重新发起。",
			res.EquipmentName, realName, req.Comment),
		models.NotifReservationRejected,
		res.ID,
	)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "已成功驳回该预约申请，已记录驳回原因并通知申请人",
		"data":    res,
	})
}

// History 获取审批流转历史记录
func (ac *ApprovalController) History(c *gin.Context) {
	type HistoryItem struct {
		models.ApprovalLog
		EquipmentName  string    `json:"equipment_name"`
		ExperimentName string    `json:"experiment_name"`
		ApplicantName  string    `json:"applicant_name"`
		GroupName      string    `json:"group_name"`
		StartTime      time.Time `json:"start_time"`
		EndTime        time.Time `json:"end_time"`
	}

	var items []HistoryItem
	database.DB.Table("approval_logs").
		Select("approval_logs.*, reservations.equipment_name, reservations.experiment_name, reservations.user_name as applicant_name, reservations.group_name, reservations.start_time, reservations.end_time").
		Joins("LEFT JOIN reservations ON reservations.id = approval_logs.reservation_id").
		Order("approval_logs.created_at DESC").
		Limit(100).
		Scan(&items)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": items,
	})
}
