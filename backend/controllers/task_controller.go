package controllers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"lab-reservation-system/database"
	"lab-reservation-system/models"
	"lab-reservation-system/services"
)

type TaskController struct{}

func (tc *TaskController) Preview(c *gin.Context) {
	var req services.TaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "请填写有效的工序和时间窗口"})
		return
	}
	plan, err := services.PlanTask(database.DB, c.GetUint("user_id"), req, false, time.Now())
	if err != nil {
		c.JSON(400, gin.H{"code": 400, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": plan})
}

func (tc *TaskController) Create(c *gin.Context) {
	var req services.TaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "请填写有效的任务信息"})
		return
	}
	task, err := services.SubmitTask(database.DB, c.GetUint("user_id"), req)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, services.ErrScheduleChanged) {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"code": status, "message": err.Error()})
		return
	}
	// Notify approvers after committing; each step follows the existing approval policy.
	for _, r := range task.Reservations {
		if r.Status == models.ReservationPendingAdvisor {
			services.DefaultNotificationService.NotifyGroupAdvisors(r.GroupID, "实验任务工序待审批", task.Name+" · "+r.ExperimentName, models.NotifReservationPending, r.ID)
		}
		if r.Status == models.ReservationPendingAdmin {
			services.DefaultNotificationService.NotifyAdmins("实验任务工序待审批", task.Name+" · "+r.ExperimentName, models.NotifReservationPending, r.ID)
		}
	}
	c.JSON(200, gin.H{"code": 200, "message": "实验任务已提交，各工序将按设备规则审批", "data": task})
}

func (tc *TaskController) List(c *gin.Context) {
	tasks := []models.ExperimentTask{}
	err := database.DB.Where("user_id = ?", c.GetUint("user_id")).Preload("Reservations", func(db *gorm.DB) *gorm.DB { return db.Order("task_step ASC") }).Order("id DESC").Find(&tasks).Error
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "读取实验任务失败"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "data": tasks})
}
