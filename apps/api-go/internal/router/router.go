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
	engine.Use(gin.Logger(), gin.Recovery(), middleware.CORS(), middleware.CurrentUser(deps.CookieName))

	engine.GET("/api/health", deps.HealthController.Health)
	engine.GET("/api/auth/feishu/login", deps.AuthController.Login)
	engine.GET("/api/auth/feishu/callback", deps.AuthController.Callback)
	engine.GET("/api/me", deps.AuthController.Me)

	engine.GET("/api/sessions", deps.SessionController.ListSessions)
	engine.POST("/api/sessions", deps.SessionController.CreateSession)
	engine.GET("/api/sessions/:sessionID", deps.SessionController.GetSession)
	engine.POST("/api/sessions/:sessionID/messages", deps.SessionController.AddMessage)
	engine.POST("/api/sessions/:sessionID/publish", deps.SessionController.Publish)

	engine.GET("/api/tasks/:taskID", deps.TaskController.GetTask)
	engine.PATCH("/api/tasks/:taskID/status", deps.TaskController.UpdateTaskStatus)

	adminGroup := engine.Group("/api/admin")
	adminGroup.Use(middleware.AdminOnly(deps.AuthService))
	adminGroup.POST("/role-mappings", deps.AdminController.CreateRoleMapping)
	adminGroup.POST("/knowledge/sync", deps.AdminController.SyncKnowledge)

	return engine
}
