package feishu

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"feishu-pipeline/apps/api-go/internal/model"
	"feishu-pipeline/apps/api-go/internal/utils"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

const (
	grantTypeAuthorizationCode = "authorization_code"
	openAPIAppAccessTokenPath  = "/open-apis/auth/v3/app_access_token/internal"
	openAPIUserTokenPath       = "/open-apis/authen/v1/access_token"
	openAPIRefreshTokenPath    = "/open-apis/authen/v1/refresh_access_token"
	openAPIUserInfoPath        = "/open-apis/authen/v1/user_info"
)

type SendResult struct {
	Channel    string
	Receiver   string
	Status     string
	RemoteID   string
	RawPayload string
}

type UserToken struct {
	AccessToken           string
	RefreshToken          string
	AccessTokenExpiresAt  time.Time
	RefreshTokenExpiresAt time.Time
}

type UserProfile struct {
	OpenID          string
	UnionID         string
	FeishuUserID    string
	Name            string
	EnName          string
	Email           string
	EnterpriseEmail string
	AvatarURL       string
}

type Config struct {
	Enabled         bool
	AppID           string
	AppSecret       string
	RedirectURL     string
	OpenBaseURL     string
	BotName         string
	ReceiveIDType   string
	BitableAppToken string
	BitableTableID  string
	BaseURL         string
}

type Client struct {
	cfg        Config
	sdk        *lark.Client
	httpClient *http.Client
}

func NewClient(cfg Config) *Client {
	return &Client{
		cfg: cfg,
		sdk: lark.NewClient(cfg.AppID, cfg.AppSecret),
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *Client) Enabled() bool {
	return c.cfg.Enabled && c.cfg.AppID != "" && c.cfg.AppSecret != ""
}

func (c *Client) AppID() string {
	return c.cfg.AppID
}

func (c *Client) GetAppAccessToken(ctx context.Context) (string, error) {
	if !c.Enabled() {
		return "", errors.New("feishu sso is not enabled")
	}

	body, err := json.Marshal(map[string]string{
		"app_id":     c.cfg.AppID,
		"app_secret": c.cfg.AppSecret,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.openAPIURL(openAPIAppAccessTokenPath), bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	var response struct {
		Code           int    `json:"code"`
		Msg            string `json:"msg"`
		AppAccessToken string `json:"app_access_token"`
	}
	if err := c.doJSON(req, &response); err != nil {
		return "", err
	}
	if response.Code != 0 {
		return "", fmt.Errorf("get feishu app_access_token failed: %s", response.Msg)
	}
	if strings.TrimSpace(response.AppAccessToken) == "" {
		return "", errors.New("feishu app_access_token is empty")
	}
	return response.AppAccessToken, nil
}

func (c *Client) ExchangeCodeForUserToken(ctx context.Context, code string) (UserToken, error) {
	if !c.Enabled() {
		return UserToken{}, errors.New("feishu sso is not enabled")
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return UserToken{}, errors.New("feishu pre-auth code is required")
	}

	appAccessToken, err := c.GetAppAccessToken(ctx)
	if err != nil {
		return UserToken{}, err
	}

	return c.requestUserToken(ctx, openAPIUserTokenPath, appAccessToken, map[string]string{
		"grant_type": grantTypeAuthorizationCode,
		"code":       code,
	})
}

func (c *Client) RefreshUserToken(ctx context.Context, refreshToken string) (UserToken, error) {
	if !c.Enabled() {
		return UserToken{}, errors.New("feishu sso is not enabled")
	}
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return UserToken{}, errors.New("feishu refresh_token is required")
	}

	appAccessToken, err := c.GetAppAccessToken(ctx)
	if err != nil {
		return UserToken{}, err
	}

	return c.requestUserToken(ctx, openAPIRefreshTokenPath, appAccessToken, map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
	})
}

func (c *Client) GetUserInfo(ctx context.Context, userAccessToken string) (UserProfile, error) {
	userAccessToken = strings.TrimSpace(userAccessToken)
	if userAccessToken == "" {
		return UserProfile{}, errors.New("feishu user_access_token is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.openAPIURL(openAPIUserInfoPath), nil)
	if err != nil {
		return UserProfile{}, err
	}
	req.Header.Set("Authorization", "Bearer "+userAccessToken)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	var response struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			OpenID          string `json:"open_id"`
			UnionID         string `json:"union_id"`
			UserID          string `json:"user_id"`
			Name            string `json:"name"`
			EnName          string `json:"en_name"`
			Email           string `json:"email"`
			EnterpriseEmail string `json:"enterprise_email"`
			AvatarURL       string `json:"avatar_url"`
		} `json:"data"`
	}
	if err := c.doJSON(req, &response); err != nil {
		return UserProfile{}, err
	}
	if response.Code != 0 {
		return UserProfile{}, fmt.Errorf("get feishu user info failed: %s", response.Msg)
	}

	return UserProfile{
		OpenID:          response.Data.OpenID,
		UnionID:         response.Data.UnionID,
		FeishuUserID:    response.Data.UserID,
		Name:            response.Data.Name,
		EnName:          response.Data.EnName,
		Email:           response.Data.Email,
		EnterpriseEmail: response.Data.EnterpriseEmail,
		AvatarURL:       response.Data.AvatarURL,
	}, nil
}

func (c *Client) CreateTaskDoc(ctx context.Context, sessionTitle string, task model.Task) (string, error) {
	if !c.Enabled() {
		return fmt.Sprintf("%s/mock/feishu/docs/%s", c.cfg.BaseURL, task.ID), nil
	}

	// 当前阶段仍保留可运行的开发回退；后续补齐真实云文档字段映射时，只需要替换这里的实现。
	return fmt.Sprintf("%s/mock/feishu/docs/%s?title=%s", c.cfg.BaseURL, task.ID, url.QueryEscape(sessionTitle)), nil
}

func (c *Client) UpsertTaskRecord(ctx context.Context, task model.Task) (string, error) {
	if !c.Enabled() || c.cfg.BitableAppToken == "" || c.cfg.BitableTableID == "" {
		return fmt.Sprintf("%s/mock/feishu/bitable/%s", c.cfg.BaseURL, task.ID), nil
	}

	// 多维表格需要业务侧配置 app token / table id；未配置时保持 mock，不阻塞主流程。
	return fmt.Sprintf("%s/mock/feishu/bitable/%s", c.cfg.BaseURL, task.ID), nil
}

func (c *Client) SendTaskMessage(ctx context.Context, task model.Task) (SendResult, error) {
	receiveID := task.AssigneeName
	content := fmt.Sprintf(`{"text":"您有一条新的需求任务：%s\n文档：%s"}`, task.Title, utils.Coalesce(task.DocURL, "待生成"))

	if !c.Enabled() {
		return SendResult{
			Channel:    "feishu-bot",
			Receiver:   receiveID,
			Status:     "mock_sent",
			RemoteID:   "mock_" + task.ID,
			RawPayload: content,
		}, nil
	}

	resp, err := c.sdk.Im.Message.Create(ctx, larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(c.cfg.ReceiveIDType).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(receiveID).
			MsgType("text").
			Content(content).
			Build()).
		Build())
	if err != nil {
		return SendResult{}, err
	}
	if !resp.Success() {
		return SendResult{}, fmt.Errorf("send feishu message failed: code=%d msg=%s", resp.Code, resp.Msg)
	}

	return SendResult{
		Channel:    "feishu-bot",
		Receiver:   receiveID,
		Status:     "accepted",
		RemoteID:   stringValue(resp.Data.MessageId),
		RawPayload: content,
	}, nil
}

func (c *Client) requestUserToken(ctx context.Context, path string, appAccessToken string, payload map[string]string) (UserToken, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return UserToken{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.openAPIURL(path), bytes.NewReader(body))
	if err != nil {
		return UserToken{}, err
	}
	req.Header.Set("Authorization", "Bearer "+appAccessToken)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	var response struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			AccessToken      string `json:"access_token"`
			RefreshToken     string `json:"refresh_token"`
			ExpiresIn        int64  `json:"expires_in"`
			RefreshExpiresIn int64  `json:"refresh_expires_in"`
		} `json:"data"`
	}
	if err := c.doJSON(req, &response); err != nil {
		return UserToken{}, err
	}
	if response.Code != 0 {
		return UserToken{}, fmt.Errorf("request feishu user token failed: %s", response.Msg)
	}

	now := time.Now().UTC()
	return UserToken{
		AccessToken:           response.Data.AccessToken,
		RefreshToken:          response.Data.RefreshToken,
		AccessTokenExpiresAt:  now.Add(time.Duration(response.Data.ExpiresIn) * time.Second),
		RefreshTokenExpiresAt: now.Add(time.Duration(response.Data.RefreshExpiresIn) * time.Second),
	}, nil
}

func (c *Client) openAPIURL(path string) string {
	base := strings.TrimRight(utils.Coalesce(c.cfg.OpenBaseURL, "https://open.feishu.cn"), "/")
	return base + path
}

func (c *Client) doJSON(req *http.Request, target any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("feishu api request failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return fmt.Errorf("decode feishu api response: %w", err)
	}
	return nil
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
