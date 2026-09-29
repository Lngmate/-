package controllers

import (
	"net/http"
	"strconv"

	"lab-reservation-system/database"
	"lab-reservation-system/models"

	"github.com/gin-gonic/gin"
)

type NotificationController struct{}

// List 获取当前登录用户的站内消息列表
func (nc *NotificationController) List(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}
	userID := userIDVal.(uint)

	var list []models.Notification
	var unreadCount int64

	// 查询未读数量
	database.DB.Model(&models.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&unreadCount)

	// 查询最近 50 条消息
	database.DB.Where("user_id = ?", userID).
		Order("is_read ASC, created_at DESC").
		Limit(50).
		Find(&list)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":         list,
			"unread_count": unreadCount,
		},
	})
}

// UnreadCount 仅获取未读消息数（用于顶部铃铛红点高效轮询）
func (nc *NotificationController) UnreadCount(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未登录"})
		return
	}
	userID := userIDVal.(uint)

	var unreadCount int64
	database.DB.Model(&models.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&unreadCount)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"unread_count": unreadCount,
		},
	})
}

// MarkRead 标记单条消息为已读
func (nc *NotificationController) MarkRead(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "消息ID无效"})
		return
	}

	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uint)

	database.DB.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "已标记为已读",
	})
}

// MarkAllRead 一键将当前用户的所有消息标为已读
func (nc *NotificationController) MarkAllRead(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uint)

	database.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "全部消息已标记为已读",
	})
}
