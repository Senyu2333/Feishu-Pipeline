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
	"sync"
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
	Enabled            bool
	AppID              string
	AppSecret          string
	RedirectURL        string
	OpenBaseURL        string
	BotName            string
	ReceiveIDType      string
	DocFolderToken     string
	BitableName        string
	BitableFolderToken string
	BitableAppToken    string
	BitableTableID     string
	BaseURL            string
}

type TaskRecordResult struct {
	AppToken  string
	TableID   string
	RecordID  string
	RecordURL string
}

type bitableTarget struct {
	AppToken string
	TableID  string
}

type Client struct {
	cfg        Config
	sdk        *lark.Client
	httpClient *http.Client
	mu         sync.Mutex
	bitable    *bitableTarget
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
	if !c.Enabled() || c.cfg.DocFolderToken == "" {
		return fmt.Sprintf("%s/mock/feishu/docs/%s", c.cfg.BaseURL, task.ID), nil
	}

	appAccessToken, err := c.GetAppAccessToken(ctx)
	if err != nil {
		return "", err
	}

	var response struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Document struct {
				DocumentID string `json:"document_id"`
			} `json:"document"`
		} `json:"data"`
	}

	if err := c.doJSONWithToken(ctx, http.MethodPost, "/open-apis/docx/v1/documents", appAccessToken, map[string]any{
		"folder_token": c.cfg.DocFolderToken,
		"title":        fmt.Sprintf("%s-%s", sessionTitle, task.Title),
	}, &response); err != nil {
		return "", err
	}
	if response.Code != 0 {
		return "", fmt.Errorf("create feishu doc failed: %s", response.Msg)
	}
	if response.Data.Document.DocumentID == "" {
		return "", errors.New("feishu document id is empty")
	}

	return fmt.Sprintf("https://feishu.cn/docx/%s", url.PathEscape(response.Data.Document.DocumentID)), nil
}

func (c *Client) UpsertTaskRecord(ctx context.Context, task model.Task) (TaskRecordResult, error) {
	target, enabled, err := c.ensureBitableTarget(ctx, task)
	if err != nil {
		return TaskRecordResult{}, err
	}
	if !enabled {
		return TaskRecordResult{
			AppToken:  task.BitableAppToken,
			TableID:   task.BitableTableID,
			RecordID:  task.BitableRecordID,
			RecordURL: fmt.Sprintf("%s/mock/feishu/bitable/%s", c.cfg.BaseURL, task.ID),
		}, nil
	}

	appAccessToken, err := c.GetAppAccessToken(ctx)
	if err != nil {
		return TaskRecordResult{}, err
	}

	fields := c.buildTaskRecordFields(task)
	recordID := strings.TrimSpace(task.BitableRecordID)
	if recordID == "" {
		var createResp struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
			Data struct {
				Record struct {
					RecordID string `json:"record_id"`
				} `json:"record"`
				RecordID string `json:"record_id"`
			} `json:"data"`
		}

		if err := c.doJSONWithToken(ctx, http.MethodPost, fmt.Sprintf("/open-apis/bitable/v1/apps/%s/tables/%s/records", target.AppToken, target.TableID), appAccessToken, map[string]any{
			"fields": fields,
		}, &createResp); err != nil {
			return TaskRecordResult{}, err
		}
		if createResp.Code != 0 {
			return TaskRecordResult{}, fmt.Errorf("create bitable record failed: %s", createResp.Msg)
		}
		recordID = utils.Coalesce(createResp.Data.Record.RecordID, createResp.Data.RecordID)
	} else {
		var updateResp struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		}

		if err := c.doJSONWithToken(ctx, http.MethodPut, fmt.Sprintf("/open-apis/bitable/v1/apps/%s/tables/%s/records/%s", target.AppToken, target.TableID, recordID), appAccessToken, map[string]any{
			"fields": fields,
		}, &updateResp); err != nil {
			return TaskRecordResult{}, err
		}
		if updateResp.Code != 0 {
			return TaskRecordResult{}, fmt.Errorf("update bitable record failed: %s", updateResp.Msg)
		}
	}

	return TaskRecordResult{
		AppToken:  target.AppToken,
		TableID:   target.TableID,
		RecordID:  recordID,
		RecordURL: fmt.Sprintf("https://feishu.cn/base/%s?table=%s", url.QueryEscape(target.AppToken), url.QueryEscape(target.TableID)),
	}, nil
}

func (c *Client) SendTaskMessage(ctx context.Context, task model.Task) (SendResult, error) {
	receiveID := strings.TrimSpace(task.AssigneeID)
	if receiveID == "" {
		receiveID = task.AssigneeName
	}

	messageText := strings.TrimSpace(task.NotifyContent)
	if messageText == "" {
		messageText = fmt.Sprintf("您有一条新的需求任务：%s", task.Title)
	}
	messageText = fmt.Sprintf("%s\n文档：%s\n排期：%s - %s",
		messageText,
		utils.Coalesce(task.DocURL, "待生成"),
		formatTaskDate(task.PlannedStartAt),
		formatTaskDate(task.PlannedEndAt),
	)
	payload, _ := json.Marshal(map[string]string{"text": messageText})
	content := string(payload)

	if !c.Enabled() || strings.TrimSpace(task.AssigneeID) == "" {
		return SendResult{
			Channel:    "feishu-bot",
			Receiver:   receiveID,
			Status:     "mock_sent",
			RemoteID:   "mock_" + task.ID,
			RawPayload: content,
		}, nil
	}

	resp, err := c.sdk.Im.Message.Create(ctx, larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(utils.Coalesce(strings.TrimSpace(task.AssigneeIDType), c.cfg.ReceiveIDType)).
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

func (c *Client) ensureBitableTarget(ctx context.Context, task model.Task) (bitableTarget, bool, error) {
	if !c.Enabled() {
		return bitableTarget{}, false, nil
	}
	if task.BitableAppToken != "" && task.BitableTableID != "" {
		return bitableTarget{AppToken: task.BitableAppToken, TableID: task.BitableTableID}, true, nil
	}
	if c.cfg.BitableAppToken != "" && c.cfg.BitableTableID != "" {
		return bitableTarget{AppToken: c.cfg.BitableAppToken, TableID: c.cfg.BitableTableID}, true, nil
	}
	if c.cfg.BitableFolderToken == "" {
		return bitableTarget{}, false, nil
	}

	c.mu.Lock()
	if c.bitable != nil {
		target := *c.bitable
		c.mu.Unlock()
		return target, true, nil
	}
	c.mu.Unlock()

	appAccessToken, err := c.GetAppAccessToken(ctx)
	if err != nil {
		return bitableTarget{}, false, err
	}

	var createAppResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			App struct {
				AppToken string `json:"app_token"`
			} `json:"app"`
			AppToken string `json:"app_token"`
		} `json:"data"`
	}

	if err := c.doJSONWithToken(ctx, http.MethodPost, "/open-apis/bitable/v1/apps", appAccessToken, map[string]any{
		"name":         utils.Coalesce(c.cfg.BitableName, "需求排期"),
		"folder_token": c.cfg.BitableFolderToken,
		"time_zone":    "Asia/Shanghai",
	}, &createAppResp); err != nil {
		return bitableTarget{}, false, err
	}
	if createAppResp.Code != 0 {
		return bitableTarget{}, false, fmt.Errorf("create bitable app failed: %s", createAppResp.Msg)
	}

	appToken := utils.Coalesce(createAppResp.Data.App.AppToken, createAppResp.Data.AppToken)
	if appToken == "" {
		return bitableTarget{}, false, errors.New("bitable app token is empty")
	}

	tableID, err := c.ensureBitableTable(ctx, appAccessToken, appToken)
	if err != nil {
		return bitableTarget{}, false, err
	}

	target := bitableTarget{AppToken: appToken, TableID: tableID}
	c.mu.Lock()
	c.bitable = &target
	c.mu.Unlock()
	return target, true, nil
}

func (c *Client) ensureBitableTable(ctx context.Context, appAccessToken string, appToken string) (string, error) {
	var listResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Items []struct {
				TableID string `json:"table_id"`
				Name    string `json:"name"`
			} `json:"items"`
		} `json:"data"`
	}

	if err := c.doJSONWithToken(ctx, http.MethodGet, fmt.Sprintf("/open-apis/bitable/v1/apps/%s/tables", appToken), appAccessToken, nil, &listResp); err != nil {
		return "", err
	}
	if listResp.Code != 0 {
		return "", fmt.Errorf("list bitable tables failed: %s", listResp.Msg)
	}

	for _, item := range listResp.Data.Items {
		if item.Name == utils.Coalesce(c.cfg.BitableName, "需求排期") {
			if err := c.ensureBitableFields(ctx, appAccessToken, appToken, item.TableID); err != nil {
				return "", err
			}
			return item.TableID, nil
		}
	}

	var createResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Table struct {
				TableID string `json:"table_id"`
			} `json:"table"`
		} `json:"data"`
	}

	if err := c.doJSONWithToken(ctx, http.MethodPost, fmt.Sprintf("/open-apis/bitable/v1/apps/%s/tables", appToken), appAccessToken, map[string]any{
		"table": map[string]any{
			"name": utils.Coalesce(c.cfg.BitableName, "需求排期"),
		},
	}, &createResp); err != nil {
		return "", err
	}
	if createResp.Code != 0 {
		return "", fmt.Errorf("create bitable table failed: %s", createResp.Msg)
	}
	if createResp.Data.Table.TableID == "" {
		return "", errors.New("bitable table id is empty")
	}

	if err := c.ensureBitableFields(ctx, appAccessToken, appToken, createResp.Data.Table.TableID); err != nil {
		return "", err
	}
	return createResp.Data.Table.TableID, nil
}

func (c *Client) ensureBitableFields(ctx context.Context, appAccessToken string, appToken string, tableID string) error {
	var listResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Items []struct {
				FieldName string `json:"field_name"`
			} `json:"items"`
		} `json:"data"`
	}

	if err := c.doJSONWithToken(ctx, http.MethodGet, fmt.Sprintf("/open-apis/bitable/v1/apps/%s/tables/%s/fields", appToken, tableID), appAccessToken, nil, &listResp); err != nil {
		return err
	}
	if listResp.Code != 0 {
		return fmt.Errorf("list bitable fields failed: %s", listResp.Msg)
	}

	existing := make(map[string]struct{}, len(listResp.Data.Items))
	for _, item := range listResp.Data.Items {
		existing[item.FieldName] = struct{}{}
	}

	for _, field := range []struct {
		Name string
		Type int
	}{
		{Name: "需求ID", Type: 1},
		{Name: "需求标题", Type: 1},
		{Name: "任务ID", Type: 1},
		{Name: "任务标题", Type: 1},
		{Name: "任务类型", Type: 1},
		{Name: "负责人", Type: 1},
		{Name: "负责人ID", Type: 1},
		{Name: "优先级", Type: 1},
		{Name: "预计工期", Type: 2},
		{Name: "计划开始", Type: 5},
		{Name: "计划结束", Type: 5},
		{Name: "状态", Type: 1},
		{Name: "通知文案", Type: 1},
		{Name: "文档链接", Type: 1},
	} {
		if _, ok := existing[field.Name]; ok {
			continue
		}

		var createResp struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		}
		if err := c.doJSONWithToken(ctx, http.MethodPost, fmt.Sprintf("/open-apis/bitable/v1/apps/%s/tables/%s/fields", appToken, tableID), appAccessToken, map[string]any{
			"field_name": field.Name,
			"type":       field.Type,
		}, &createResp); err != nil {
			return err
		}
		if createResp.Code != 0 {
			return fmt.Errorf("create bitable field %s failed: %s", field.Name, createResp.Msg)
		}
	}

	return nil
}

func (c *Client) buildTaskRecordFields(task model.Task) map[string]any {
	fields := map[string]any{
		"需求ID":  task.SessionID,
		"需求标题":  task.Title,
		"任务ID":  task.ID,
		"任务标题":  task.Title,
		"任务类型":  string(task.Type),
		"负责人":   task.AssigneeName,
		"负责人ID": task.AssigneeID,
		"优先级":   string(task.Priority),
		"预计工期":  task.EstimateDays,
		"状态":    string(task.Status),
		"通知文案":  task.NotifyContent,
		"文档链接":  task.DocURL,
	}
	if task.PlannedStartAt != nil {
		fields["计划开始"] = task.PlannedStartAt.UnixMilli()
	}
	if task.PlannedEndAt != nil {
		fields["计划结束"] = task.PlannedEndAt.UnixMilli()
	}
	return fields
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

func (c *Client) doJSONWithToken(ctx context.Context, method string, path string, token string, payload any, target any) error {
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.openAPIURL(path), body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	return c.doJSON(req, target)
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func formatTaskDate(value *time.Time) string {
	if value == nil {
		return "待排期"
	}
	return value.Format("2006-01-02")
}
