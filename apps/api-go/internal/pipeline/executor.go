package pipeline

import (
	"context"
	"encoding/json"
	"time"

	"feishu-pipeline/apps/api-go/internal/model"
)

type StageExecutionResult struct {
	ArtifactType model.ArtifactType
	Title        string
	ContentText  string
	ContentJSON  string
	OutputJSON   string
	AgentRun     *AgentObservation
}

type Executor interface {
	Execute(context.Context, StageContext) (StageExecutionResult, error)
}

type StageContext struct {
	Run         model.PipelineRun
	Stage       model.StageRun
	Artifacts   []model.Artifact
	Checkpoints []model.Checkpoint
	Input       map[string]any
	AgentRunID  string // 预创建的AgentRun ID，为空则表示需要创建新的
}

func baseStagePayload(ctx StageContext) map[string]any {
	return map[string]any{
		"runId":       ctx.Run.ID,
		"stageKey":    ctx.Stage.StageKey,
		"generatedAt": time.Now().UTC().Format(time.RFC3339),
		"attempt":     ctx.Stage.Attempt,
		"input":       ctx.Input,
	}
}

func newStageResult(artifactType model.ArtifactType, title string, payload map[string]any, contentText string) StageExecutionResult {
	contentJSON, _ := json.Marshal(payload)
	return StageExecutionResult{ArtifactType: artifactType, Title: title, ContentText: contentText, ContentJSON: string(contentJSON), OutputJSON: string(contentJSON)}
}
