package services

import (
	"time"

	"lab-reservation-system/database"
	"lab-reservation-system/models"
)

type NotificationService struct{}

var DefaultNotificationService = &NotificationService{}

// SendNotification 向单个用户发送站内消息
func (ns *NotificationService) SendNotification(userID uint, title, content string, notifType models.NotificationType, relatedID uint) {
	if userID == 0 {
		return
	}
	notif := models.Notification{
		UserID:    userID,
		Title:     title,
		Content:   content,
		Type:      notifType,
		RelatedID: relatedID,
		IsRead:    false,
		CreatedAt: time.Now(),
	}
	database.DB.Create(&notif)
}

// NotifyAdmins 向所有设备管理员和系统管理员广播消息
func (ns *NotificationService) NotifyAdmins(title, content string, notifType models.NotificationType, relatedID uint) {
	var admins []models.User
	database.DB.Where("role IN ?", []models.Role{models.RoleEquipmentAdmin, models.RoleSystemAdmin}).Find(&admins)
	for _, admin := range admins {
		ns.SendNotification(admin.ID, title, content, notifType, relatedID)
	}
}

// NotifyGroupAdvisors 向指定课题组内的指导导师发送消息
func (ns *NotificationService) NotifyGroupAdvisors(groupID uint, title, content string, notifType models.NotificationType, relatedID uint) {
	var advisors []models.User
	database.DB.Where("group_id = ? AND role = ?", groupID, models.RoleAdvisor).Find(&advisors)
	for _, adv := range advisors {
		ns.SendNotification(adv.ID, title, content, notifType, relatedID)
	}
}
