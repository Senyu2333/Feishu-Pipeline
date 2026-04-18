package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"feishu-pipeline/apps/api-go/internal/external/feishu"
	"feishu-pipeline/apps/api-go/internal/model"
)

type PipelineService struct {
	feishuClient *feishu.Client
}

func NewPipelineService(feishuClient *feishu.Client) *PipelineService {
	return &PipelineService{feishuClient: feishuClient}
}

type PipelineCreateResult struct {
	TableURL   string   `json:"tableUrl"`
	RecordURLs []string `json:"recordUrls"`
	AppToken   string   `json:"appToken"`
	TableID    string   `json:"tableId"`
	RecordIDs  []string `json:"recordIds"`
}

func (s *PipelineService) CreatePipeline(ctx context.Context, tasks []model.Task) (*PipelineCreateResult, error) {
	templateToken := s.feishuClient.BitableTemplateToken()
	if templateToken == "" {
		return nil, fmt.Errorf("bitable_template_token is not configured")
	}

	newAppToken, err := s.feishuClient.CopyBitableTemplate(ctx, templateToken)
	if err != nil {
		return nil, fmt.Errorf("copy template: %w", err)
	}
	log.Printf("[pipeline] copied bitable template: new_app_token=%s", newAppToken)

	var tableID string
	for i := range 8 {
		time.Sleep(time.Duration(i+1) * 2 * time.Second)
		tableID, err = s.feishuClient.GetBitableTableID(ctx, newAppToken)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("get table id after retries: %w", err)
	}
	log.Printf("[pipeline] got table_id=%s for app_token=%s", tableID, newAppToken)

	now := time.Now()
	records := make([]feishu.BitableRecord, 0, len(tasks))
	for _, t := range tasks {
		fields := map[string]any{
			"需求":   t.Title,
			"负责人":  t.AssigneeName,
			"任务类型": string(t.Type),
			"优先级":  string(t.Priority),
			"状态":   "未开始",
			"开始时间": now.UnixMilli(),
		}
		if t.PlannedEndAt != nil {
			fields["截止时间"] = t.PlannedEndAt.UnixMilli()
		} else {
			fields["截止时间"] = now.AddDate(0, 0, int(t.EstimateDays)).UnixMilli()
		}
		if t.DocURL != "" {
			fields["文档链接"] = t.DocURL
		}
		records = append(records, feishu.BitableRecord{Fields: fields})
	}

	recordIDs, err := s.feishuClient.BatchCreateBitableRecords(ctx, newAppToken, tableID, records)
	if err != nil {
		return nil, fmt.Errorf("batch create records: %w", err)
	}
	log.Printf("[pipeline] created %d bitable records", len(recordIDs))

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
