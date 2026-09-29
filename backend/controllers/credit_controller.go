package controllers

import (
	"net/http"
	"strconv"

	"lab-reservation-system/database"
	"lab-reservation-system/models"
	"lab-reservation-system/services"

	"github.com/gin-gonic/gin"
)

type CreditController struct{}

// Detail 获取用户科研信誉详情与履约评级
func (cc *CreditController) Detail(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uint)

	// 如果传入了 target_user_id 且当前为管理员，可查任意用户
	targetIDStr := c.Query("user_id")
	if targetIDStr != "" {
		if tid, err := strconv.Atoi(targetIDStr); err == nil {
			userID = uint(tid)
		}
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	var records []models.CreditRecord
	database.DB.Where("user_id = ?", user.ID).Order("created_at DESC").Find(&records)

	// 评级计算
	rating := "卓越"
	badgeColor := "success"
	if user.CreditScore >= 95 {
		rating = "五星卓越 (享有高峰期优先排队与免导师审批特权)"
		badgeColor = "success"
	} else if user.CreditScore >= 80 {
		rating = "良好守诺 (标准预约权限)"
		badgeColor = "primary"
	} else if user.CreditScore >= 60 {
		rating = "预警关注 (限制一级精密设备预约)"
		badgeColor = "warning"
	} else {
		rating = "失信冻结 (账号暂停预约，需线下重置)"
		badgeColor = "danger"
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"user_id":      user.ID,
			"user_name":    user.RealName,
			"credit_score": user.CreditScore,
			"is_frozen":    user.IsFrozen,
			"rating":       rating,
			"badge_color":  badgeColor,
			"history":      records,
		},
	})
}

// Adjust 手动调整用户信誉积分（管理员权限）
func (cc *CreditController) Adjust(c *gin.Context) {
	type AdjustReq struct {
		UserID int    `json:"user_id" binding:"required"`
		Delta  int    `json:"delta" binding:"required"`
		Reason string `json:"reason" binding:"required"`
	}

	var req AdjustReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	newScore, err := services.DefaultCreditEngine.AdjustCreditScore(nil, uint(req.UserID), 0, req.Delta, req.Reason, "admin_adjustment")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "调整失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      200,
		"message":   "信誉积分调整成功",
		"new_score": newScore,
	})
}
