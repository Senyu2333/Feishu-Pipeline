package router

import (
	"feishu-pipeline/apps/api-go/internal/controller"
	"feishu-pipeline/apps/api-go/internal/middleware"
	"feishu-pipeline/apps/api-go/internal/service"

	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	CookieName         string
	HealthController   *controller.HealthController
	AuthController     *controller.AuthController
	SessionController  *controller.SessionController
	TaskController     *controller.TaskController
	AdminController    *controller.AdminController
	PipelineController *controller.PipelineController
	AuthService        *service.AuthService
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
	authenticated.POST("/sessions/:sessionID/messages/stream", deps.SessionController.StreamMessage)
	authenticated.POST("/sessions/:sessionID/publish", deps.SessionController.Publish)
	authenticated.POST("/sessions/:sessionID/auto-publish-check", deps.SessionController.AutoPublishCheck)
	authenticated.GET("/tasks/:taskID", deps.TaskController.GetTask)
	authenticated.PATCH("/tasks/:taskID/status", deps.TaskController.UpdateTaskStatus)

	adminGroup := authenticated.Group("/admin")
	adminGroup.Use(middleware.AdminOnly(deps.AuthService))
	adminGroup.POST("/role-mappings", deps.AdminController.CreateRoleMapping)
	adminGroup.GET("/role-owners", deps.AdminController.ListRoleOwners)
	adminGroup.POST("/role-owners", deps.AdminController.SaveRoleOwner)
	adminGroup.POST("/knowledge/sync", deps.AdminController.SyncKnowledge)

	authenticated.GET("/pipeline-templates", deps.PipelineController.ListTemplates)
	authenticated.GET("/pipeline-runs", deps.PipelineController.ListRuns)
	authenticated.POST("/pipeline-runs", deps.PipelineController.CreateRun)
	authenticated.POST("/pipeline-runs/from-session", deps.PipelineController.CreateRunFromSession)
	authenticated.GET("/pipeline-runs/:id", deps.PipelineController.GetRun)
	authenticated.GET("/pipeline-runs/:id/timeline", deps.PipelineController.GetRunTimeline)
	authenticated.GET("/pipeline-runs/:id/current", deps.PipelineController.GetRunCurrent)
	authenticated.GET("/pipeline-runs/:id/stages", deps.PipelineController.ListStages)
	authenticated.GET("/pipeline-runs/:id/artifacts", deps.PipelineController.ListArtifacts)
	authenticated.GET("/pipeline-runs/:id/checkpoints", deps.PipelineController.ListCheckpoints)
	authenticated.GET("/pipeline-runs/:id/agent-runs", deps.PipelineController.ListAgentRuns)
	authenticated.GET("/pipeline-runs/:id/deliveries", deps.PipelineController.ListGitDeliveries)
	authenticated.POST("/pipeline-runs/:id/start", deps.PipelineController.StartRun)
	authenticated.POST("/pipeline-runs/:id/pause", deps.PipelineController.PauseRun)
	authenticated.POST("/pipeline-runs/:id/resume", deps.PipelineController.ResumeRun)
	authenticated.POST("/pipeline-runs/:id/terminate", deps.PipelineController.TerminateRun)
	authenticated.GET("/git-deliveries/:deliveryID", deps.PipelineController.GetGitDelivery)
	authenticated.POST("/checkpoints/:checkpointID/approve", deps.PipelineController.ApproveCheckpoint)
	authenticated.POST("/checkpoints/:checkpointID/reject", deps.PipelineController.RejectCheckpoint)

	return engine
}
