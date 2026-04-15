package middleware

import (
	"errors"
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
		ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		ctx.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")

		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}
		ctx.Next()
	}
}

func CurrentUser(authService *service.AuthService, cookieName string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		currentUserID := ""
		if cookieValue, err := ctx.Cookie(cookieName); err == nil && cookieValue != "" {
			resolvedUserID, resolveErr := authService.ResolveSessionUserID(ctx.Request.Context(), cookieValue)
			switch {
			case resolveErr == nil:
				currentUserID = resolvedUserID
			case errors.Is(resolveErr, service.ErrAuthenticationRequired):
			default:
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, commontype.Envelope{Error: resolveErr.Error()})
				return
			}
		}

		ctx.Set(CurrentUserIDKey(), currentUserID)
		ctx.Next()
	}
}

func RequireAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		currentUserID, _ := ctx.Get(CurrentUserIDKey())
		if userID, ok := currentUserID.(string); !ok || userID == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, commontype.Envelope{Error: service.ErrAuthenticationRequired.Error()})
			return
		}
		ctx.Next()
	}
}

func AdminOnly(authService *service.AuthService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		currentUserID, _ := ctx.Get(CurrentUserIDKey())
		user, err := authService.CurrentUser(ctx.Request.Context(), currentUserID.(string))
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, commontype.Envelope{Error: err.Error()})
			return
		}
		if user.Role != model.RoleAdmin {
			ctx.AbortWithStatusJSON(http.StatusForbidden, commontype.Envelope{Error: "admin permission required"})
			return
		}
		ctx.Next()
	}
}
