package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"lab-reservation-system/database"
	"lab-reservation-system/models"

	"github.com/gin-gonic/gin"
)

type GroupController struct{}

// List 获取课题组列表
func (gc *GroupController) List(c *gin.Context) {
	var groups []models.ResearchGroup
	if err := database.DB.Order("id ASC").Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取课题组失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": groups,
	})
}

// Detail 获取课题组详情与成员
func (gc *GroupController) Detail(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var group models.ResearchGroup
	if err := database.DB.First(&group, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "课题组不存在"})
		return
	}

	var members []models.User
	database.DB.Where("group_id = ?", id).Find(&members)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"group":   group,
			"members": members,
		},
	})
}

// Create 新增课题组
func (gc *GroupController) Create(c *gin.Context) {
	type CreateGroupReq struct {
		Name       string  `json:"name" binding:"required"`
		LeaderName string  `json:"leader_name" binding:"required"`
		Department string  `json:"department" binding:"required"`
		Balance    float64 `json:"balance"`
	}

	var req CreateGroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请填写完整的课题组信息"})
		return
	}

	group := models.ResearchGroup{
		Name:           strings.TrimSpace(req.Name),
		LeaderName:     strings.TrimSpace(req.LeaderName),
		Department:     strings.TrimSpace(req.Department),
		Balance:        req.Balance,
		TotalHoursUsed: 0,
	}

	if err := database.DB.Create(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "新增课题组失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "新增课题组成功",
		"data":    group,
	})
}

// Update 修改课题组信息
func (gc *GroupController) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var group models.ResearchGroup
	if err := database.DB.First(&group, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "课题组不存在"})
		return
	}

	type UpdateGroupReq struct {
		Name       string `json:"name"`
		LeaderName string `json:"leader_name"`
		Department string `json:"department"`
	}

	var req UpdateGroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	if req.Name != "" {
		group.Name = strings.TrimSpace(req.Name)
	}
	if req.LeaderName != "" {
		group.LeaderName = strings.TrimSpace(req.LeaderName)
	}
	if req.Department != "" {
		group.Department = strings.TrimSpace(req.Department)
	}

	database.DB.Save(&group)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "课题组信息更新成功",
		"data":    group,
	})
}

// Delete 删除课题组
func (gc *GroupController) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var group models.ResearchGroup
	if err := database.DB.First(&group, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "课题组不存在"})
		return
	}

	var memberCount int64
	database.DB.Model(&models.User{}).Where("group_id = ?", id).Count(&memberCount)
	if memberCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": fmt.Sprintf("该课题组下仍有 %d 位归属成员，请先调整成员归属后再删除", memberCount),
		})
		return
	}

	database.DB.Delete(&group)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "课题组已成功删除",
	})
}

// Transactions 获取课题组虚拟结算消费与流水
func (gc *GroupController) Transactions(c *gin.Context) {
	groupID := c.Query("group_id")

	query := database.DB.Model(&models.TransactionRecord{})
	if groupID != "" {
		query = query.Where("group_id = ?", groupID)
	}

	var records []models.TransactionRecord
	query.Order("created_at DESC").Find(&records)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": records,
	})
}

// Recharge 虚拟科研经费充值/划拨
func (gc *GroupController) Recharge(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	type RechargeReq struct {
		Amount float64 `json:"amount" binding:"required"`
		Desc   string  `json:"desc"`
	}

	var req RechargeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请输入充值金额"})
		return
	}

	var group models.ResearchGroup
	if err := database.DB.First(&group, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "课题组不存在"})
		return
	}

	group.Balance += req.Amount
	database.DB.Save(&group)

	record := models.TransactionRecord{
		GroupID:      group.ID,
		GroupName:    group.Name,
		Amount:       req.Amount,
		BalanceAfter: group.Balance,
		Type:         "recharge",
		Description:  req.Desc,
		CreatedAt:    time.Now(),
	}
	if record.Description == "" {
		record.Description = "学院科研机时补贴与专项经费划转"
	}
	database.DB.Create(&record)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "充值/划拨成功！",
		"data":    group,
	})
}
