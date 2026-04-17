package service

import (
	"context"
	"fmt"
	"time"

	"feishu-pipeline/apps/api-go/internal/external/feishu"
)

const (
	templateAppToken = "A7krw99hJiatrZktgmzcOVmkn6d"
)

// typeToP 任务类型 → 优先级
var typeToP = map[string]string{
	"frontend": "P0",
	"backend":  "P1",
	"shared":   "P2",
}

// PipelineTask 对应 AI 输出的 Task 结构
type PipelineTask struct {
	ID                 string   `json:"id"`
	SessionID          string   `json:"sessionId"`
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	Type               string   `json:"type"`
	Status             string   `json:"status"`
	AssigneeName       string   `json:"assigneeName"`
	AssigneeRole       string   `json:"assigneeRole"`
	AcceptanceCriteria []string `json:"acceptanceCriteria"`
	Risks              []string `json:"risks"`
}

// PipelineRequirement 对应 AI 输出的 Requirement 结构
type PipelineRequirement struct {
	ID                  string   `json:"id"`
	SessionID           string   `json:"sessionId"`
	Title               string   `json:"title"`
	Summary             string   `json:"summary"`
	Status              string   `json:"status"`
	DeliverySummary     string   `json:"deliverySummary"`
	ReferencedKnowledge []string `json:"referencedKnowledge"`
	PublishedAt         string   `json:"publishedAt"`
}

// PipelineResult AI 流水线输出
type PipelineResult struct {
	Requirement PipelineRequirement `json:"requirement"`
	Tasks       []PipelineTask      `json:"tasks"`
}

// PipelineCreateResult 接口返回
type PipelineCreateResult struct {
	TableURL   string   `json:"tableUrl"`
	RecordURLs []string `json:"recordUrls"`
	AppToken   string   `json:"appToken"`
	TableID    string   `json:"tableId"`
	RecordIDs  []string `json:"recordIds"`
}

// mockTasks 当 AI 未输出时使用的 mock 数据（对应 TaskMock.js）
var mockTasks = []PipelineTask{
	{
		ID:           "recvh0ReNieoZ8",
		Title:        "AI生成前端任务",
		Type:         "frontend",
		Status:       "todo",
		AssigneeName: "ou_2910013f1e6456f16a0ce75ede950a0a",
		AssigneeRole: "frontend",
	},
	{
		ID:           "recvh0ReNikCDG",
		Title:        "AI生成后端任务",
		Type:         "backend",
		Status:       "todo",
		AssigneeName: "ou_e04138c9633dd0d2ea166d79f548ab5d",
		AssigneeRole: "backend",
	},
}

type PipelineService struct {
	feishuClient *feishu.Client
}

func NewPipelineService(feishuClient *feishu.Client) *PipelineService {
	return &PipelineService{feishuClient: feishuClient}
}

func (s *PipelineService) CreatePipeline(ctx context.Context, result PipelineResult) (*PipelineCreateResult, error) {
	tasks := result.Tasks
	if len(tasks) == 0 {
		tasks = mockTasks
	}

	// 步骤 1：复制模板表格，得到新的 app_token
	newAppToken, err := s.feishuClient.CopyBitableTemplate(ctx, templateAppToken)
	if err != nil {
		return nil, fmt.Errorf("copy template: %w", err)
	}

	// 步骤 2：获取新表格的第一个 table_id（飞书复制模板是异步的，需重试等待）
	var tableID string
	for i := range 8 {
		time.Sleep(time.Duration(i+1) * 2 * time.Second)
		tableID, err = s.feishuClient.GetBitableTableID(ctx, newAppToken)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("get table id: %w", err)
	}

	// 步骤 3：构造记录并批量写入
	now := time.Now()
	deadline := now.AddDate(0, 0, 7)
	nowMs := now.UnixMilli()
	deadlineMs := deadline.UnixMilli()

	records := make([]feishu.BitableRecord, 0, len(tasks))
	for _, t := range tasks {
		priority := typeToP[t.Type]
		if priority == "" {
			priority = "P3"
		}

		fields := map[string]any{
			"需求":   t.Title,
			"优先级":  priority,
			"状态":   "未开始",
			"开始时间": nowMs,
			"截止时间": deadlineMs,
		}

		records = append(records, feishu.BitableRecord{Fields: fields})
	}

	recordIDs, err := s.feishuClient.BatchCreateBitableRecords(ctx, newAppToken, tableID, records)
	if err != nil {
		return nil, fmt.Errorf("batch create records: %w", err)
	}

	// 拼接 URL
	tableURL := fmt.Sprintf("https://feishu.cn/base/%s?table=%s", newAppToken, tableID)
	recordURLs := make([]string, len(recordIDs))
	for i, rid := range recordIDs {
		recordURLs[i] = fmt.Sprintf("%s&record=%s", tableURL, rid)
	}

	return &PipelineCreateResult{
		TableURL:   tableURL,
		RecordURLs: recordURLs,
		AppToken:   newAppToken,
		TableID:    tableID,
		RecordIDs:  recordIDs,
	}, nil
}
