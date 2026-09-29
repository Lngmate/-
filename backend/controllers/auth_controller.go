package controllers

import (
	"net/http"
	"strings"

	"lab-reservation-system/database"
	"lab-reservation-system/middleware"
	"lab-reservation-system/models"
	"lab-reservation-system/utils"

	"github.com/gin-gonic/gin"
)

type AuthController struct{}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string      `json:"username" binding:"required"`
	Password string      `json:"password" binding:"required"`
	RealName string      `json:"real_name" binding:"required"`
	Email    string      `json:"email"`
	Phone    string      `json:"phone"`
	Role     models.Role `json:"role"`
	GroupID  *uint       `json:"group_id"`
}

type QuickSwitchRequest struct {
	Role models.Role `json:"role" binding:"required"`
}

type UpdateProfileRequest struct {
	RealName string `json:"real_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// Register 用户注册
func (ac *AuthController) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请填写完整的注册信息"})
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.RealName = strings.TrimSpace(req.RealName)

	if len(req.Username) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户名长度不能少于3个字符"})
		return
	}
	if len(req.Password) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "密码长度不能少于3个字符"})
		return
	}

	// 检查用户名是否已存在
	var count int64
	database.DB.Model(&models.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "该用户名已存在，请使用其他用户名"})
		return
	}

	// 密码安全加密
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		hashedPassword = req.Password
	}

	role := req.Role
	if role == "" {
		role = models.RoleStudent
	}
	// 不允许直接注册为系统管理员
	if role == models.RoleSystemAdmin {
		role = models.RoleEquipmentAdmin
	}

	var groupName string
	if req.GroupID != nil && *req.GroupID > 0 {
		var group models.ResearchGroup
		if err := database.DB.First(&group, *req.GroupID).Error; err == nil {
			groupName = group.Name
		}
	}

	user := models.User{
		Username:    req.Username,
		Password:    hashedPassword,
		RealName:    req.RealName,
		Email:       req.Email,
		Phone:       req.Phone,
		Role:        role,
		GroupID:     req.GroupID,
		GroupName:   groupName,
		CreditScore: 100,
		IsFrozen:    false,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "用户注册失败: " + err.Error()})
		return
	}

	// 生成 Token
	token, err := middleware.GenerateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "注册成功但生成令牌失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "恭喜，用户注册成功！",
		"data": gin.H{
			"token": token,
			"user":  user,
		},
	})
}

// Login 用户登录
func (ac *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请输入账号和密码"})
		return
	}

	var user models.User
	if err := database.DB.Where("username = ?", strings.TrimSpace(req.Username)).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户名或密码错误"})
		return
	}

	if !utils.CheckPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户名或密码错误"})
		return
	}

	if user.IsFrozen {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您的账号已被系统管理员冻结，请联系管理中心核实解冻"})
		return
	}

	token, err := middleware.GenerateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "生成Token失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登录成功，欢迎进入科研设备共享系统",
		"data": gin.H{
			"token": token,
			"user":  user,
		},
	})
}

// QuickSwitch 快捷角色切换（答辩演示与测试体验核心功能）
func (ac *AuthController) QuickSwitch(c *gin.Context) {
	var req QuickSwitchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请指定目标角色"})
		return
	}

	var targetUser models.User
	if err := database.DB.Where("role = ?", req.Role).First(&targetUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "该角色暂无预置账号"})
		return
	}

	token, err := middleware.GenerateToken(&targetUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "切换生成Token失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "角色切换成功",
		"data": gin.H{
			"token": token,
			"user":  targetUser,
		},
	})
}

// Profile 获取当前登录用户信息
func (ac *AuthController) Profile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": user,
	})
}

// UpdateProfile 更新个人资料
func (ac *AuthController) UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数格式错误"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	if req.RealName != "" {
		user.RealName = strings.TrimSpace(req.RealName)
	}
	user.Email = strings.TrimSpace(req.Email)
	user.Phone = strings.TrimSpace(req.Phone)

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新资料失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "个人资料更新成功",
		"data":    user,
	})
}

// ChangePassword 修改密码
func (ac *AuthController) ChangePassword(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请输入原密码与新密码"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	if !utils.CheckPassword(req.OldPassword, user.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "原密码不正确"})
		return
	}

	if len(req.NewPassword) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "新密码长度不能少于3位"})
		return
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		hashedPassword = req.NewPassword
	}

	user.Password = hashedPassword
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "密码修改失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "密码修改成功，请牢记新密码",
	})
}

// Logout 退出登录
func (ac *AuthController) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "已安全退出系统",
	})
}
