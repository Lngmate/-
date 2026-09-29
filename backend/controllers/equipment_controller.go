package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"lab-reservation-system/database"
	"lab-reservation-system/models"

	"github.com/gin-gonic/gin"
)

type EquipmentController struct{}

// List 获取设备台账列表
func (ec *EquipmentController) List(c *gin.Context) {
	category := c.Query("category")
	labRoom := c.Query("lab_room")
	status := c.Query("status")
	keyword := c.Query("keyword")

	query := database.DB.Model(&models.Equipment{})

	if category != "" {
		query = query.Where("category = ?", category)
	}
	if labRoom != "" {
		query = query.Where("lab_room = ?", labRoom)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		pattern := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR model LIKE ? OR location LIKE ? OR specs LIKE ?", pattern, pattern, pattern, pattern)
	}

	var equipments []models.Equipment
	if err := query.Order("id ASC").Find(&equipments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取设备列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": equipments,
	})
}

// Detail 获取设备详细信息
func (ec *EquipmentController) Detail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "设备ID无效"})
		return
	}

	var equipment models.Equipment
	if err := database.DB.First(&equipment, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "设备不存在"})
		return
	}

	// 查关联维保记录
	var maintenanceRecords []models.MaintenanceRecord
	database.DB.Where("equipment_id = ?", id).Order("created_at DESC").Find(&maintenanceRecords)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"equipment":   equipment,
			"maintenance": maintenanceRecords,
		},
	})
}

// Create 新增设备
func (ec *EquipmentController) Create(c *gin.Context) {
	var equipment models.Equipment
	if err := c.ShouldBindJSON(&equipment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数格式错误"})
		return
	}

	if equipment.HealthScore == 0 {
		equipment.HealthScore = 100
	}
	if equipment.Status == "" {
		equipment.Status = models.StatusAvailable
	}

	if err := database.DB.Create(&equipment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建设备失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "新增设备成功",
		"data":    equipment,
	})
}

// Update 更新设备信息
func (ec *EquipmentController) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "设备ID无效"})
		return
	}

	var equipment models.Equipment
	if err := database.DB.First(&equipment, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "设备未找到"})
		return
	}

	var req models.Equipment
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数格式错误"})
		return
	}

	req.ID = uint(id)
	if err := database.DB.Save(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新设备失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "设备信息更新成功",
		"data":    req,
	})
}

// Delete 删除设备
func (ec *EquipmentController) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "设备ID无效"})
		return
	}

	var equipment models.Equipment
	if err := database.DB.First(&equipment, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "设备未找到"})
		return
	}

	// 检查是否存在未完成的预约
	var activeCount int64
	database.DB.Model(&models.Reservation{}).
		Where("equipment_id = ? AND status IN ('pending_advisor', 'pending_admin', 'approved', 'active')", id).
		Count(&activeCount)

	if activeCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": fmt.Sprintf("该设备尚有 %d 笔进行中或审批中的预约，无法直接删除！请先完结或取消预约。", activeCount),
		})
		return
	}

	database.DB.Delete(&equipment)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "设备【" + equipment.Name + "】已成功从台账中移除",
	})
}

// TriggerMaintenance 触发维保与健康度调整（创新点：健康度与自适应维保）
func (ec *EquipmentController) TriggerMaintenance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "设备ID无效"})
		return
	}

	type MaintenanceReq struct {
		Type        string  `json:"type" binding:"required"` // "routine", "repair", "calibration"
		Description string  `json:"description"`
		Cost        float64 `json:"cost"`
	}

	var req MaintenanceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数不全"})
		return
	}

	var equipment models.Equipment
	if err := database.DB.First(&equipment, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "设备不存在"})
		return
	}

	userName, _ := c.Get("real_name")
	userID, _ := c.Get("user_id")

	scoreBefore := equipment.HealthScore
	scoreAfter := 100 // 维保完成后健康度恢复至100

	record := models.MaintenanceRecord{
		EquipmentID:       equipment.ID,
		EquipmentName:     equipment.Name,
		OperatorID:        userID.(uint),
		OperatorName:      userName.(string),
		Type:              req.Type,
		Description:       req.Description,
		Cost:              req.Cost,
		HealthScoreBefore: scoreBefore,
		HealthScoreAfter:  scoreAfter,
		StartTime:         time.Now(),
		EndTime:           time.Now().Add(2 * time.Hour),
		CreatedAt:         time.Now(),
	}

	database.DB.Create(&record)

	// 更新设备状态与健康度
	equipment.HealthScore = scoreAfter
	equipment.Status = models.StatusAvailable
	database.DB.Save(&equipment)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "维保完成，设备健康指数已重置为100分，状态恢复空闲就绪！",
		"data":    equipment,
	})
}
