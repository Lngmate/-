package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"lab-reservation-system/database"
	"lab-reservation-system/models"
	"lab-reservation-system/utils"

	"github.com/gin-gonic/gin"
)

type UserController struct{}

// List 获取用户列表（支持关键词、角色、课题组、冻结状态筛选）
func (uc *UserController) List(c *gin.Context) {
	keyword := c.Query("keyword")
	role := c.Query("role")
	groupIDStr := c.Query("group_id")
	isFrozenStr := c.Query("is_frozen")

	query := database.DB.Model(&models.User{})

	if keyword != "" {
		pattern := "%" + strings.TrimSpace(keyword) + "%"
		query = query.Where("username LIKE ? OR real_name LIKE ? OR email LIKE ? OR phone LIKE ?", pattern, pattern, pattern, pattern)
	}
	if role != "" {
		query = query.Where("role = ?", role)
	}
	if groupIDStr != "" {
		if gid, err := strconv.Atoi(groupIDStr); err == nil {
			query = query.Where("group_id = ?", gid)
		}
	}
	if isFrozenStr != "" {
		if isFrozenStr == "true" {
			query = query.Where("is_frozen = ?", true)
		} else if isFrozenStr == "false" {
			query = query.Where("is_frozen = ?", false)
		}
	}

	var users []models.User
	if err := query.Order("id ASC").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": users,
	})
}

// Detail 获取单个用户详情
func (uc *UserController) Detail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID无效"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": user,
	})
}

// Update 管理员编辑用户信息（角色、所属课题组、真实姓名等）
func (uc *UserController) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID无效"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	type UpdateReq struct {
		RealName string      `json:"real_name"`
		Role     models.Role `json:"role"`
		GroupID  *uint       `json:"group_id"`
		Email    string      `json:"email"`
		Phone    string      `json:"phone"`
	}

	var req UpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数格式错误"})
		return
	}

	if req.RealName != "" {
		user.RealName = strings.TrimSpace(req.RealName)
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	user.Email = strings.TrimSpace(req.Email)
	user.Phone = strings.TrimSpace(req.Phone)

	if req.GroupID != nil {
		user.GroupID = req.GroupID
		if *req.GroupID > 0 {
			var group models.ResearchGroup
			if err := database.DB.First(&group, *req.GroupID).Error; err == nil {
				user.GroupName = group.Name
			}
		} else {
			user.GroupID = nil
			user.GroupName = ""
		}
	}

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新用户信息失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "用户信息更新成功",
		"data":    user,
	})
}

// ToggleFreeze 快速切换用户冻结/解冻状态
func (uc *UserController) ToggleFreeze(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID无效"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	// 禁止冻结主系统管理员
	if user.Username == "sysadmin" {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "系统保护：禁止冻结顶级系统管理员账号"})
		return
	}

	user.IsFrozen = !user.IsFrozen
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "操作失败"})
		return
	}

	actionStr := "解冻并恢复正常"
	if user.IsFrozen {
		actionStr = "已冻结限制登录与预约"
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "用户【" + user.RealName + "】" + actionStr,
		"data":    user,
	})
}

// ResetPassword 管理员重置用户密码为 123456
func (uc *UserController) ResetPassword(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID无效"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	hashedPassword, _ := utils.HashPassword("123456")
	user.Password = hashedPassword

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "密码重置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "用户【" + user.RealName + "】的密码已成功重置为 123456",
	})
}

// Delete 管理员删除用户
func (uc *UserController) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID无效"})
		return
	}

	currentUserID, _ := c.Get("user_id")
	if uint(id) == currentUserID.(uint) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无法删除当前正在登录的管理员账号"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	if user.Username == "sysadmin" {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "核心系统管理员不可删除"})
		return
	}

	database.DB.Delete(&user)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "用户【" + user.RealName + "】已成功删除",
	})
}
