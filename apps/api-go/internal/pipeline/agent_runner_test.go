package pipeline_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"feishu-pipeline/apps/api-go/internal/model"
	"feishu-pipeline/apps/api-go/internal/pipeline"
)

type stubProvider struct {
	name    string
	model   string
	content string
	err     error
}

func (p stubProvider) Name() string {
	return p.name
}

func (p stubProvider) Model() string {
	return p.model
}

func (p stubProvider) Generate(context.Context, pipeline.AgentProviderRequest) (pipeline.AgentProviderResponse, error) {
	return pipeline.AgentProviderResponse{Content: p.content, LatencyMS: 7, TokenUsage: pipeline.TokenUsage{InputTokens: 10, OutputTokens: 20, TotalTokens: 30}}, p.err
}

func TestAgentRunnerUsesProviderJSONOutput(t *testing.T) {
	provider := stubProvider{name: "stub", model: "stub-model", content: `{"summary":"provider summary","goals":["g1"],"acceptanceCriteria":["a1"]}`}
	executor := pipeline.NewSequentialExecutor(pipeline.WithAgentRunner(pipeline.NewAgentRunner(provider, pipeline.DefaultPromptRegistry())))

	result, err := executor.Execute(context.Background(), stageContext(pipeline.StageRequirementAnalysis, model.StageTypeAnalysis, map[string]any{}))
	if err != nil {
		t.Fatalf("execute with provider: %v", err)
	}
	if result.AgentRun == nil {
		t.Fatalf("expected agent observation")
	}
	if result.AgentRun.Provider != "stub" || result.AgentRun.Model != "stub-model" {
		t.Fatalf("unexpected provider observation: %+v", result.AgentRun)
	}
	if result.AgentRun.LatencyMS != 7 {
		t.Fatalf("expected provider latency, got %d", result.AgentRun.LatencyMS)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(result.OutputJSON), &payload); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if payload["summary"] != "provider summary" {
		t.Fatalf("expected provider summary, got %v", payload["summary"])
	}
	if !strings.Contains(result.AgentRun.PromptSnapshot, pipeline.AgentRequirementAnalyst) {
		t.Fatalf("expected prompt snapshot to include agent key")
	}
}

func TestAgentRunnerFallsBackWhenProviderOutputInvalid(t *testing.T) {
	provider := stubProvider{name: "stub", model: "stub-model", content: `{"summary":""}`}
	executor := pipeline.NewSequentialExecutor(pipeline.WithAgentRunner(pipeline.NewAgentRunner(provider, pipeline.DefaultPromptRegistry())))

	result, err := executor.Execute(context.Background(), stageContext(pipeline.StageRequirementAnalysis, model.StageTypeAnalysis, map[string]any{}))
	if err != nil {
		t.Fatalf("execute with invalid provider output: %v", err)
	}
	if result.AgentRun == nil {
		t.Fatalf("expected agent observation")
	}
	if result.AgentRun.Provider != pipeline.AgentProviderDeterministic {
		t.Fatalf("expected deterministic fallback, got %s", result.AgentRun.Provider)
	}
	if !strings.Contains(result.AgentRun.TokenUsageJSON, "schema_validation_error") {
		t.Fatalf("expected fallback reason in token usage: %s", result.AgentRun.TokenUsageJSON)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(result.OutputJSON), &payload); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if strings.TrimSpace(payload["summary"].(string)) == "" {
		t.Fatalf("expected deterministic fallback summary")
	}
}

func TestAgentRunnerFallsBackWhenProviderReturnsError(t *testing.T) {
	provider := stubProvider{name: "stub", model: "stub-model", err: errors.New("provider timeout")}
	executor := pipeline.NewSequentialExecutor(pipeline.WithAgentRunner(pipeline.NewAgentRunner(provider, pipeline.DefaultPromptRegistry())))

	result, err := executor.Execute(context.Background(), stageContext(pipeline.StageRequirementAnalysis, model.StageTypeAnalysis, map[string]any{}))
	if err != nil {
		t.Fatalf("execute with provider error: %v", err)
	}
	if result.AgentRun == nil {
		t.Fatalf("expected agent observation")
	}
	if result.AgentRun.Provider != pipeline.AgentProviderDeterministic {
		t.Fatalf("expected deterministic fallback, got %s", result.AgentRun.Provider)
	}
	if !strings.Contains(result.AgentRun.TokenUsageJSON, "provider_error") {
		t.Fatalf("expected provider error fallback reason in token usage: %s", result.AgentRun.TokenUsageJSON)
	}
}
