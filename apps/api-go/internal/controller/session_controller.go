package controller

import (
	"net/http"

	"feishu-pipeline/apps/api-go/internal/repo"
	"feishu-pipeline/apps/api-go/internal/service"
	sessiontype "feishu-pipeline/apps/api-go/internal/type/session"
	tasktype "feishu-pipeline/apps/api-go/internal/type/task"

	"github.com/gin-gonic/gin"
)

type SessionController struct {
	sessionService *service.SessionService
	publishService *service.PublishService
}

func NewSessionController(sessionService *service.SessionService, publishService *service.PublishService) *SessionController {
	return &SessionController{
		sessionService: sessionService,
		publishService: publishService,
	}
}

// ListSessions
// @tags 会话
// @summary 会话列表
// @router /api/sessions [GET]
// @produce application/json
func (c *SessionController) ListSessions(ctx *gin.Context) {
	items, err := c.sessionService.ListSessions(ctx.Request.Context())
	if err != nil {
		writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	response := make([]sessiontype.SessionSummaryResponse, 0, len(items))
	for _, item := range items {
		response = append(response, sessiontype.NewSessionSummaryResponse(item.Session, item.OwnerName, item.MessageCount))
	}
	writeSuccess(ctx, http.StatusOK, response)
}

// CreateSession
// @tags 会话
// @summary 创建会话
// @router /api/sessions [POST]
// @accept application/json
// @produce application/json
// @param req body sessiontype.CreateSessionRequest true "json入参"
func (c *SessionController) CreateSession(ctx *gin.Context) {
	var request sessiontype.CreateSessionRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}

	detail, err := c.sessionService.CreateSession(ctx.Request.Context(), currentUserID(ctx), request.Title, request.Prompt)
	if err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	writeSuccess(ctx, http.StatusCreated, mapSessionDetail(detail))
}

// GetSession
// @tags 会话
// @summary 会话详情
// @router /api/sessions/{sessionID} [GET]
// @param sessionID path string true "会话ID"
// @produce application/json
func (c *SessionController) GetSession(ctx *gin.Context) {
	detail, err := c.sessionService.GetSessionDetail(ctx.Request.Context(), ctx.Param("sessionID"))
	if err != nil {
		writeError(ctx, http.StatusNotFound, err)
		return
	}
	writeSuccess(ctx, http.StatusOK, mapSessionDetail(detail))
}

// AddMessage
// @tags 会话
// @summary 追加消息
// @router /api/sessions/{sessionID}/messages [POST]
// @accept application/json
// @produce application/json
// @param sessionID path string true "会话ID"
// @param req body sessiontype.CreateMessageRequest true "json入参"
func (c *SessionController) AddMessage(ctx *gin.Context) {
	var request sessiontype.CreateMessageRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}

	if err := c.sessionService.AddMessage(ctx.Request.Context(), currentUserID(ctx), ctx.Param("sessionID"), request.Content); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}

	detail, err := c.sessionService.GetSessionDetail(ctx.Request.Context(), ctx.Param("sessionID"))
	if err != nil {
		writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	writeSuccess(ctx, http.StatusCreated, mapSessionDetail(detail))
}

// Publish
// @tags 会话
// @summary 发布会话
// @router /api/sessions/{sessionID}/publish [POST]
// @produce application/json
// @param sessionID path string true "会话ID"
func (c *SessionController) Publish(ctx *gin.Context) {
	if err := c.publishService.PublishSession(ctx.Request.Context(), currentUserID(ctx), ctx.Param("sessionID")); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}
	writeSuccess(ctx, http.StatusAccepted, acceptedMessage())
}

func mapSessionDetail(aggregate *repo.SessionAggregate) sessiontype.SessionDetailResponse {
	messageResponses := make([]sessiontype.MessageResponse, 0, len(aggregate.Messages))
	for _, message := range aggregate.Messages {
		messageResponses = append(messageResponses, sessiontype.NewMessageResponse(message))
	}

	taskResponses := make([]tasktype.TaskResponse, 0, len(aggregate.Tasks))
	for _, task := range aggregate.Tasks {
		taskResponses = append(taskResponses, tasktype.NewTaskResponse(task))
	}

	return sessiontype.SessionDetailResponse{
		Session:     sessiontype.NewSessionSummaryResponse(aggregate.Session, aggregate.Owner.Name, aggregate.MessageCount),
		Messages:    messageResponses,
		Requirement: sessiontype.NewRequirementResponse(aggregate.Requirement),
		Tasks:       taskResponses,
	}
}
