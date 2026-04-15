package controller

import (
	"net/http"
	"strings"
	"time"

	"feishu-pipeline/apps/api-go/internal/model"
	"feishu-pipeline/apps/api-go/internal/service"
	authtype "feishu-pipeline/apps/api-go/internal/type/auth"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService    *service.AuthService
	cookieName     string
	cookieTTL      time.Duration
	cookieSecure   bool
	cookieSameSite http.SameSite
}

func NewAuthController(authService *service.AuthService, cookieName string, cookieTTL time.Duration, cookieSecure bool, cookieSameSite string) *AuthController {
	return &AuthController{
		authService:    authService,
		cookieName:     cookieName,
		cookieTTL:      cookieTTL,
		cookieSecure:   cookieSecure,
		cookieSameSite: parseSameSite(cookieSameSite),
	}
}

func (c *AuthController) FeishuConfig(ctx *gin.Context) {
	writeSuccess(ctx, http.StatusOK, authtype.FeishuSSOConfigResponse{
		Enabled: c.authService.FeishuEnabled(),
		AppID:   c.authService.FeishuAppID(),
	})
}

func (c *AuthController) SSOLogin(ctx *gin.Context) {
	var request authtype.FeishuSSOLoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		writeError(ctx, http.StatusBadRequest, err)
		return
	}

	user, session, err := c.authService.LoginByCode(ctx.Request.Context(), request.Code)
	if err != nil {
		writeError(ctx, http.StatusBadGateway, err)
		return
	}

	c.writeSessionCookie(ctx, session)
	writeSuccess(ctx, http.StatusOK, authtype.NewLoginResponse(user))
}

func (c *AuthController) Logout(ctx *gin.Context) {
	sessionID, _ := ctx.Cookie(c.cookieName)
	if err := c.authService.Logout(ctx.Request.Context(), sessionID); err != nil {
		writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	c.clearSessionCookie(ctx)
	writeSuccess(ctx, http.StatusOK, map[string]string{"status": "logged_out"})
}

func (c *AuthController) Me(ctx *gin.Context) {
	user, err := c.authService.CurrentUser(ctx.Request.Context(), currentUserID(ctx))
	if err != nil {
		writeError(ctx, http.StatusUnauthorized, err)
		return
	}
	writeSuccess(ctx, http.StatusOK, authtype.NewUserResponse(user))
}

func (c *AuthController) writeSessionCookie(ctx *gin.Context, session model.LoginSession) {
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     c.cookieName,
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		Secure:   c.cookieSecure,
		SameSite: c.cookieSameSite,
		MaxAge:   int(c.cookieTTL.Seconds()),
		Expires:  session.ExpiresAt,
	})
}

func (c *AuthController) clearSessionCookie(ctx *gin.Context) {
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     c.cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   c.cookieSecure,
		SameSite: c.cookieSameSite,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func parseSameSite(value string) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}
