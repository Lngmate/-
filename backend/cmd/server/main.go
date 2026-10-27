package main

import (
	"fmt"
	"log"

	"lab-reservation-system/config"
	"lab-reservation-system/controllers"
	"lab-reservation-system/database"
	"lab-reservation-system/middleware"
	"lab-reservation-system/models"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 初始化配置
	cfg := config.InitConfig()

	// 2. 初始化数据库与自动迁移
	database.InitDB(cfg)

	// 3. 设置 Gin 引擎
	r := gin.Default()

	// 全局跨域中间件
	r.Use(middleware.Cors())

	// 实例化控制器
	authCtrl := &controllers.AuthController{}
	userCtrl := &controllers.UserController{}
	eqCtrl := &controllers.EquipmentController{}
	resCtrl := &controllers.ReservationController{}
	apprCtrl := &controllers.ApprovalController{}
	grpCtrl := &controllers.GroupController{}
	credCtrl := &controllers.CreditController{}
	statCtrl := &controllers.StatsController{}
	notifCtrl := &controllers.NotificationController{}
	taskCtrl := &controllers.TaskController{}

	// API 路由定义
	api := r.Group("/api")
	{
		tasks := api.Group("/tasks").Use(middleware.AuthRequired())
		tasks.GET("", taskCtrl.List)
		tasks.POST("/preview", taskCtrl.Preview)
		tasks.POST("", taskCtrl.Create)
		// 1. 认证模块
		auth := api.Group("/auth")
		{
			auth.POST("/login", authCtrl.Login)
			auth.POST("/register", authCtrl.Register)
			auth.POST("/quick-switch", authCtrl.QuickSwitch) // 快捷切换角色（评测与答辩演示）
			auth.POST("/logout", authCtrl.Logout)
			auth.GET("/profile", middleware.AuthRequired(), authCtrl.Profile)
			auth.PUT("/profile", middleware.AuthRequired(), authCtrl.UpdateProfile)
			auth.PUT("/password", middleware.AuthRequired(), authCtrl.ChangePassword)
		}

		// 2. 用户管理模块
		users := api.Group("/users").Use(middleware.AuthRequired())
		{
			users.GET("", userCtrl.List)
			users.GET("/:id", userCtrl.Detail)
			// 管理员专属运维接口
			users.PUT("/:id", middleware.RequireRoles(models.RoleSystemAdmin), userCtrl.Update)
			users.PUT("/:id/freeze", middleware.RequireRoles(models.RoleSystemAdmin), userCtrl.ToggleFreeze)
			users.PUT("/:id/reset-pwd", middleware.RequireRoles(models.RoleSystemAdmin), userCtrl.ResetPassword)
			users.DELETE("/:id", middleware.RequireRoles(models.RoleSystemAdmin), userCtrl.Delete)
		}

		// 3. 设备台账模块
		equipments := api.Group("/equipments")
		{
			equipments.GET("", eqCtrl.List)
			equipments.GET("/:id", eqCtrl.Detail)

			// 管理员写操作
			eqAdmin := equipments.Group("").Use(middleware.AuthRequired(), middleware.RequireRoles(models.RoleEquipmentAdmin, models.RoleSystemAdmin))
			{
				eqAdmin.POST("", eqCtrl.Create)
				eqAdmin.PUT("/:id", eqCtrl.Update)
				eqAdmin.DELETE("/:id", eqCtrl.Delete)
				eqAdmin.POST("/:id/maintenance", eqCtrl.TriggerMaintenance)
			}
		}

		// 4. 预约排期与履约模块
		reservations := api.Group("/reservations").Use(middleware.AuthRequired())
		{
			reservations.GET("", resCtrl.List)
			reservations.GET("/:id", resCtrl.Detail)
			reservations.GET("/events", resCtrl.CalendarEvents)        // FullCalendar 事件流
			reservations.POST("/check", resCtrl.CheckConflictAndBatch) // 冲突与伴热拼单即时检测
			reservations.POST("", resCtrl.Create)
			reservations.POST("/:id/cancel", resCtrl.Cancel)     // 主动取消预约
			reservations.POST("/:id/checkin", resCtrl.Checkin)   // 现场扫码/核销上机
			reservations.POST("/:id/checkout", resCtrl.Checkout) // 实验结束归还与验收
		}

		// 5. 多级审批流转模块
		approvals := api.Group("/approvals").Use(middleware.AuthRequired())
		{
			approvals.GET("/pending", apprCtrl.PendingList)
			approvals.GET("/history", apprCtrl.History)
			approvals.POST("/:id/approve", apprCtrl.Approve)
			approvals.POST("/:id/reject", apprCtrl.Reject)
		}

		// 6. 课题组与虚拟结算模块
		groups := api.Group("/groups").Use(middleware.AuthRequired())
		{
			groups.GET("", grpCtrl.List)
			groups.GET("/:id", grpCtrl.Detail)
			groups.GET("/transactions", grpCtrl.Transactions)
			groups.POST("/:id/recharge", grpCtrl.Recharge)
			groups.POST("", middleware.RequireRoles(models.RoleSystemAdmin), grpCtrl.Create)
			groups.PUT("/:id", middleware.RequireRoles(models.RoleSystemAdmin, models.RoleAdvisor), grpCtrl.Update)
			groups.DELETE("/:id", middleware.RequireRoles(models.RoleSystemAdmin), grpCtrl.Delete)
		}

		// 7. 科研履约信用模块
		credits := api.Group("/credits").Use(middleware.AuthRequired())
		{
			credits.GET("/detail", credCtrl.Detail)
			credits.POST("/adjust", middleware.RequireRoles(models.RoleSystemAdmin, models.RoleEquipmentAdmin), credCtrl.Adjust)
		}

		// 8. BI 决策大屏与统计分析
		stats := api.Group("/stats").Use(middleware.AuthRequired())
		{
			stats.GET("/dashboard", statCtrl.Dashboard)
		}

		// 9. 站内消息与红点提示通知模块
		notifications := api.Group("/notifications").Use(middleware.AuthRequired())
		{
			notifications.GET("", notifCtrl.List)
			notifications.GET("/unread-count", notifCtrl.UnreadCount)
			notifications.PUT("/:id/read", notifCtrl.MarkRead)
			notifications.PUT("/read-all", notifCtrl.MarkAllRead)
		}
	}

	fmt.Printf(">>> 实验室设备预约与共享系统 API 服务正在启动，监听端口: :%s <<<\n", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("启动 HTTP 服务失败: %v", err)
	}
}
