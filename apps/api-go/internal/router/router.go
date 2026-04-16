package router

import (
	"feishu-pipeline/apps/api-go/internal/controller"
	"feishu-pipeline/apps/api-go/internal/middleware"
	"feishu-pipeline/apps/api-go/internal/service"

	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	CookieName        string
	HealthController  *controller.HealthController
	AuthController    *controller.AuthController
	SessionController *controller.SessionController
	TaskController    *controller.TaskController
	AdminController   *controller.AdminController
	AuthService       *service.AuthService
}

func New(deps Dependencies) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery(), middleware.CORS(), middleware.CurrentUser(deps.AuthService, deps.CookieName))

	engine.GET("/api/health", deps.HealthController.Health)
	engine.GET("/api/auth/feishu/config", deps.AuthController.FeishuConfig)
	engine.POST("/api/auth/feishu/sso/login", deps.AuthController.SSOLogin)
	engine.POST("/api/auth/logout", deps.AuthController.Logout)

	authenticated := engine.Group("/api")
	authenticated.Use(middleware.RequireAuth())
	authenticated.GET("/me", deps.AuthController.Me)
	authenticated.GET("/sessions", deps.SessionController.ListSessions)
	authenticated.POST("/sessions", deps.SessionController.CreateSession)
	authenticated.GET("/sessions/:sessionID", deps.SessionController.GetSession)
	authenticated.POST("/sessions/:sessionID/messages", deps.SessionController.AddMessage)
	authenticated.POST("/sessions/:sessionID/publish", deps.SessionController.Publish)
	authenticated.POST("/sessions/:sessionID/auto-publish-check", deps.SessionController.AutoPublishCheck)
	authenticated.GET("/tasks/:taskID", deps.TaskController.GetTask)
	authenticated.PATCH("/tasks/:taskID/status", deps.TaskController.UpdateTaskStatus)

	adminGroup := authenticated.Group("/admin")
	adminGroup.Use(middleware.AdminOnly(deps.AuthService))
	adminGroup.POST("/role-mappings", deps.AdminController.CreateRoleMapping)
	adminGroup.POST("/knowledge/sync", deps.AdminController.SyncKnowledge)

	return engine
}
