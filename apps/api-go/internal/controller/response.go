package controller

import (
	"net/http"

	"feishu-pipeline/apps/api-go/internal/middleware"
	commontype "feishu-pipeline/apps/api-go/internal/type/common"

	"github.com/gin-gonic/gin"
)

func writeSuccess(ctx *gin.Context, status int, data any) {
	ctx.JSON(status, commontype.Envelope{Data: data})
}

func writeError(ctx *gin.Context, status int, err error) {
	ctx.JSON(status, commontype.Envelope{Error: err.Error()})
}

func currentUserID(ctx *gin.Context) string {
	if value, ok := ctx.Get(middleware.CurrentUserIDKey()); ok {
		if userID, ok := value.(string); ok {
			return userID
		}
	}
	return ""
}

func acceptedMessage() map[string]string {
	return map[string]string{
		"status":  "accepted",
		"message": "需求已受理，后台正在生成任务和飞书分发结果。",
	}
}

const (
	statusOK       = http.StatusOK
	statusCreated  = http.StatusCreated
	statusAccepted = http.StatusAccepted
)
