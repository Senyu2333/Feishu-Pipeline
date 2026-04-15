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

	"feishu-pipeline/apps/api-go/internal/model"
	"feishu-pipeline/apps/api-go/internal/utils"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type SendResult struct {
	Channel    string
	Receiver   string
	Status     string
	RemoteID   string
	RawPayload string
}

type Config struct {
	Enabled         bool
	AppID           string
	AppSecret       string
	RedirectURL     string
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

func (c *Client) BuildLoginURL(state string) string {
	if !c.Enabled() {
		return fmt.Sprintf("%s/api/auth/feishu/callback?state=%s&mock=true", c.cfg.BaseURL, url.QueryEscape(state))
	}

	values := url.Values{}
	values.Set("app_id", c.cfg.AppID)
	values.Set("redirect_uri", c.cfg.RedirectURL)
	values.Set("state", state)
	return "https://open.feishu.cn/open-apis/authen/v1/authorize?" + values.Encode()
}

func (c *Client) ExchangeCode(ctx context.Context, code string) (model.User, error) {
	if !c.Enabled() || strings.TrimSpace(code) == "" {
		return model.User{
			ID:          "u_product_demo",
			Name:        "产品经理小明",
			Email:       "product@example.com",
			Role:        model.RoleProduct,
			Departments: []string{"产品部"},
		}, nil
	}

	payload := map[string]string{
		"grant_type": "authorization_code",
		"code":       code,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return model.User{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://open.feishu.cn/open-apis/authen/v1/oauth/token", bytes.NewReader(body))
	if err != nil {
		return model.User{}, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.SetBasicAuth(c.cfg.AppID, c.cfg.AppSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return model.User{}, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.User{}, err
	}

	var parsed struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			OpenID string `json:"open_id"`
			Name   string `json:"name"`
			Email  string `json:"email"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return model.User{}, err
	}
	if parsed.Code != 0 {
		return model.User{}, fmt.Errorf("exchange feishu oauth code failed: %s", parsed.Msg)
	}

	return model.User{
		ID:           "fs_" + parsed.Data.OpenID,
		FeishuOpenID: parsed.Data.OpenID,
		Name:         utils.Coalesce(parsed.Data.Name, "飞书用户"),
		Email:        parsed.Data.Email,
		Role:         model.RoleProduct,
		Departments:  []string{"飞书组织"},
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

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
