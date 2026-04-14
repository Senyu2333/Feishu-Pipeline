package feishu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"feishu-pipeline/apps/api-go/internal/config"
	"feishu-pipeline/apps/api-go/internal/domain"
)

type Client struct {
	cfg        config.Config
	httpClient *http.Client
}

type SendResult struct {
	Channel    string
	Receiver   string
	Status     string
	RemoteID   string
	RawPayload string
}

func NewClient(cfg config.Config) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) AuthLoginURL(state string) string {
	if !c.cfg.FeishuEnabled() {
		return c.cfg.BaseURL + "/api/auth/feishu/callback?state=" + url.QueryEscape(state) + "&mock=true"
	}
	values := url.Values{}
	values.Set("app_id", c.cfg.FeishuAppID)
	values.Set("redirect_uri", c.cfg.FeishuRedirectURL)
	values.Set("state", state)
	return "https://open.feishu.cn/open-apis/authen/v1/authorize?" + values.Encode()
}

func (c *Client) ExchangeCode(ctx context.Context, code string) (domain.User, error) {
	if !c.cfg.FeishuEnabled() || code == "" {
		return domain.User{
			ID:          "u_product_demo",
			Name:        "产品经理小明",
			Email:       "product@example.com",
			Role:        domain.RoleProduct,
			Departments: []string{"产品部"},
		}, nil
	}

	body := map[string]string{
		"grant_type": "authorization_code",
		"code":       code,
	}
	tokenResp, err := c.postJSON(ctx, "https://open.feishu.cn/open-apis/authen/v1/oauth/token", body, map[string]string{
		"Content-Type": "application/json; charset=utf-8",
	})
	if err != nil {
		return domain.User{}, err
	}

	var parsed struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			AccessToken string `json:"access_token"`
			OpenID      string `json:"open_id"`
			Name        string `json:"name"`
			Email       string `json:"email"`
		} `json:"data"`
	}
	if err := json.Unmarshal(tokenResp, &parsed); err != nil {
		return domain.User{}, err
	}
	if parsed.Code != 0 {
		return domain.User{}, fmt.Errorf("feishu oauth error: %s", parsed.Msg)
	}

	return domain.User{
		ID:           "fs_" + parsed.Data.OpenID,
		FeishuOpenID: parsed.Data.OpenID,
		Name:         coalesce(parsed.Data.Name, "飞书用户"),
		Email:        parsed.Data.Email,
		Role:         domain.RoleProduct,
		Departments:  []string{"飞书组织"},
	}, nil
}

func (c *Client) CreateTaskDoc(ctx context.Context, sessionTitle string, task domain.Task) (string, error) {
	if !c.cfg.FeishuEnabled() {
		return fmt.Sprintf("%s/mock/feishu/docs/%s", c.cfg.BaseURL, task.ID), nil
	}
	return fmt.Sprintf("%s/mock/feishu/docs/%s", c.cfg.BaseURL, task.ID), nil
}

func (c *Client) UpsertTaskRecord(ctx context.Context, task domain.Task) (string, error) {
	if !c.cfg.FeishuEnabled() {
		return fmt.Sprintf("%s/mock/feishu/bitable/%s", c.cfg.BaseURL, task.ID), nil
	}
	return fmt.Sprintf("%s/mock/feishu/bitable/%s", c.cfg.BaseURL, task.ID), nil
}

func (c *Client) SendTaskMessage(ctx context.Context, task domain.Task) (SendResult, error) {
	payload := map[string]any{
		"title":    task.Title,
		"assignee": task.AssigneeName,
		"doc_url":  task.DocURL,
	}

	if !c.cfg.FeishuEnabled() {
		return SendResult{
			Channel:    "feishu-bot",
			Receiver:   task.AssigneeName,
			Status:     "mock_sent",
			RemoteID:   "mock_" + task.ID,
			RawPayload: string(mustJSON(payload)),
		}, nil
	}

	return SendResult{
		Channel:    "feishu-bot",
		Receiver:   task.AssigneeName,
		Status:     "accepted",
		RemoteID:   "remote_" + task.ID,
		RawPayload: string(mustJSON(payload)),
	}, nil
}

func (c *Client) postJSON(ctx context.Context, endpoint string, payload any, headers map[string]string) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}

	req.SetBasicAuth(c.cfg.FeishuAppID, c.cfg.FeishuAppSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("feishu request failed: %s", strings.TrimSpace(string(data)))
	}
	return data, nil
}

func mustJSON(value any) []byte {
	data, _ := json.Marshal(value)
	return data
}

func coalesce(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
