package middleware

import (
	"net/http"

	"feishu-pipeline/apps/api-go/internal/model"
	"feishu-pipeline/apps/api-go/internal/service"
	commontype "feishu-pipeline/apps/api-go/internal/type/common"

	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		origin := ctx.GetHeader("Origin")
		if origin != "" {
			ctx.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}
		ctx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Demo-User")
		ctx.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")

		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}
		ctx.Next()
	}
}

func CurrentUser(cookieName string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		currentUserID := "u_product_demo"
		if cookieValue, err := ctx.Cookie(cookieName); err == nil && cookieValue != "" {
			currentUserID = cookieValue
		}
		if headerValue := ctx.GetHeader("X-Demo-User"); headerValue != "" {
			currentUserID = headerValue
		}

		ctx.Set(CurrentUserIDKey(), currentUserID)
		ctx.Next()
	}
}

func AdminOnly(authService *service.AuthService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		currentUserID, _ := ctx.Get(CurrentUserIDKey())
		user := authService.EnsureUser(ctx.Request.Context(), currentUserID.(string))
		if user.Role != model.RoleAdmin {
			ctx.AbortWithStatusJSON(http.StatusForbidden, commontype.Envelope{Error: "admin permission required"})
			return
		}
		ctx.Next()
	}
}
