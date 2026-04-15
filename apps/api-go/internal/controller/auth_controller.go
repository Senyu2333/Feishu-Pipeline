package controller

import (
	"net/http"

	"feishu-pipeline/apps/api-go/internal/service"
	authtype "feishu-pipeline/apps/api-go/internal/type/auth"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService *service.AuthService
	cookieName  string
}

func NewAuthController(authService *service.AuthService, cookieName string) *AuthController {
	return &AuthController{
		authService: authService,
		cookieName:  cookieName,
	}
}

func (c *AuthController) Login(ctx *gin.Context) {
	ctx.Redirect(http.StatusFound, c.authService.LoginURL("feishu-pipeline-state"))
}

func (c *AuthController) Callback(ctx *gin.Context) {
	user, err := c.authService.LoginByCode(ctx.Request.Context(), ctx.Query("code"))
	if err != nil {
		writeError(ctx, http.StatusBadGateway, err)
		return
	}

	ctx.SetCookie(c.cookieName, user.ID, 86400*7, "/", "", false, true)
	ctx.Redirect(http.StatusFound, "/")
}

func (c *AuthController) Me(ctx *gin.Context) {
	user, err := c.authService.CurrentUser(ctx.Request.Context(), currentUserID(ctx))
	if err != nil {
		writeError(ctx, http.StatusUnauthorized, err)
		return
	}
	writeSuccess(ctx, http.StatusOK, authtype.NewUserResponse(user))
}
