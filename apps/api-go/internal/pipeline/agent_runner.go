package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"feishu-pipeline/apps/api-go/internal/model"
)

type AgentRunner struct {
	provider AgentProvider
	registry *PromptRegistry
}

func NewAgentRunner(provider AgentProvider, registry *PromptRegistry) *AgentRunner {
	if registry == nil {
		registry = DefaultPromptRegistry()
	}
	return &AgentRunner{provider: provider, registry: registry}
}

func (r *AgentRunner) Execute(ctx context.Context, stageContext StageContext, fallback StageHandler) (StageExecutionResult, error) {
	spec, ok := r.registry.Build(stageContext)
	if !ok {
		return fallback.Execute(ctx, stageContext)
	}
	inputJSON := marshalJSONString(stageContext.Input)
	if r.provider == nil {
		return r.executeFallback(ctx, stageContext, fallback, spec, inputJSON, "provider_unavailable", nil)
	}

	startedAt := time.Now()
	response, err := r.provider.Generate(ctx, AgentProviderRequest{
		AgentKey:     spec.AgentKey,
		StageKey:     spec.StageKey,
		SystemPrompt: spec.SystemPrompt,
		UserPrompt:   spec.UserPrompt,
		InputJSON:    inputJSON,
		Metadata:     map[string]string{"runId": stageContext.Run.ID, "stageRunId": stageContext.Stage.ID},
	})
	latencyMS := response.LatencyMS
	if latencyMS <= 0 {
		latencyMS = time.Since(startedAt).Milliseconds()
	}
	if err != nil {
		return r.executeFallback(ctx, stageContext, fallback, spec, inputJSON, "provider_error: "+err.Error(), &response)
	}

	payload, parseErr := decodeAgentJSON(response.Content)
	if parseErr != nil {
		return r.executeFallback(ctx, stageContext, fallback, spec, inputJSON, "json_parse_error: "+parseErr.Error(), &response)
	}
	if validationErr := validateAgentPayload(payload, spec.RequiredFields); validationErr != nil {
		return r.executeFallback(ctx, stageContext, fallback, spec, inputJSON, "schema_validation_error: "+validationErr.Error(), &response)
	}

	enriched, handlerErr := r.enrichProviderPayload(ctx, stageContext, fallback, spec, payload)
	if handlerErr != nil {
		return r.executeFallback(ctx, stageContext, fallback, spec, inputJSON, "handler_post_process_error: "+handlerErr.Error(), &response)
	}
	outputJSON := marshalJSONString(enriched)
	result := StageExecutionResult{
		ArtifactType: spec.ArtifactType,
		Title:        spec.ArtifactTitle,
		ContentText:  contentTextFromPayload(spec.ArtifactTitle, enriched),
		ContentJSON:  outputJSON,
		OutputJSON:   outputJSON,
	}
	result.AgentRun = &AgentObservation{
		AgentKey:       spec.AgentKey,
		Provider:       r.provider.Name(),
		Model:          r.provider.Model(),
		PromptSnapshot: promptSnapshot(spec),
		InputJSON:      inputJSON,
		OutputJSON:     result.OutputJSON,
		TokenUsageJSON: tokenUsageJSON(response, ""),
		LatencyMS:      latencyMS,
		Status:         model.AgentRunSucceeded,
	}
	return result, nil
}

func (r *AgentRunner) enrichProviderPayload(ctx context.Context, stageContext StageContext, fallback StageHandler, spec AgentPromptSpec, providerPayload map[string]any) (map[string]any, error) {
	enriched := baseStagePayload(stageContext)
	if spec.StageKey == StageTestGeneration {
		fallbackResult, err := fallback.Execute(ctx, stageContext)
		if err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(fallbackResult.OutputJSON), &enriched)
	}
	for key, value := range providerPayload {
		if spec.StageKey == StageTestGeneration && isBackendControlledTestField(key) {
			continue
		}
		enriched[key] = value
	}
	return enriched, nil
}

func (r *AgentRunner) executeFallback(ctx context.Context, stageContext StageContext, fallback StageHandler, spec AgentPromptSpec, inputJSON string, reason string, providerResponse *AgentProviderResponse) (StageExecutionResult, error) {
	startedAt := time.Now()
	result, err := fallback.Execute(ctx, stageContext)
	status := model.AgentRunSucceeded
	errorMessage := ""
	if err != nil {
		status = model.AgentRunFailed
		errorMessage = err.Error()
	}
	usage := AgentProviderResponse{}
	if providerResponse != nil {
		usage = *providerResponse
	}
	result.AgentRun = &AgentObservation{
		AgentKey:       spec.AgentKey,
		Provider:       AgentProviderDeterministic,
		Model:          AgentModelDeterministic,
		PromptSnapshot: promptSnapshot(spec),
		InputJSON:      inputJSON,
		OutputJSON:     result.OutputJSON,
		TokenUsageJSON: tokenUsageJSON(usage, reason),
		LatencyMS:      time.Since(startedAt).Milliseconds(),
		Status:         status,
		ErrorMessage:   errorMessage,
	}
	return result, err
}

func decodeAgentJSON(content string) (map[string]any, error) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return nil, fmt.Errorf("empty provider content")
	}
	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```json")
		trimmed = strings.TrimPrefix(trimmed, "```")
		trimmed = strings.TrimSuffix(trimmed, "```")
		trimmed = strings.TrimSpace(trimmed)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func validateAgentPayload(payload map[string]any, requiredFields []string) error {
	for _, field := range requiredFields {
		value, ok := payload[field]
		if !ok {
			return fmt.Errorf("missing required field %q", field)
		}
		if value == nil {
			return fmt.Errorf("required field %q is null", field)
		}
		if text, ok := value.(string); ok && strings.TrimSpace(text) == "" {
			return fmt.Errorf("required field %q is empty", field)
		}
		if items, ok := value.([]any); ok && len(items) == 0 {
			return fmt.Errorf("required field %q is empty", field)
		}
	}
	return nil
}

func isBackendControlledTestField(key string) bool {
	switch key {
	case SchemaFieldCommands, SchemaFieldCommandResults, SchemaFieldStatus:
		return true
	default:
		return false
	}
}

func contentTextFromPayload(title string, payload map[string]any) string {
	summary, _ := payload[SchemaFieldSummary].(string)
	if strings.TrimSpace(summary) == "" {
		return title
	}
	return fmt.Sprintf("%s：%s", title, summary)
}

func promptSnapshot(spec AgentPromptSpec) string {
	return marshalJSONString(map[string]any{
		"agentKey":     spec.AgentKey,
		"stageKey":     spec.StageKey,
		"systemPrompt": spec.SystemPrompt,
		"userPrompt":   spec.UserPrompt,
	})
}

func tokenUsageJSON(response AgentProviderResponse, fallbackReason string) string {
	payload := map[string]any{
		"inputTokens":  response.TokenUsage.InputTokens,
		"outputTokens": response.TokenUsage.OutputTokens,
		"totalTokens":  response.TokenUsage.TotalTokens,
		"finishReason": response.FinishReason,
		"raw":          response.TokenUsage.Raw,
		"metadata":     response.Metadata,
	}
	if strings.TrimSpace(response.RawJSON) != "" {
		payload["rawResponse"] = response.RawJSON
	}
	if strings.TrimSpace(fallbackReason) != "" {
		payload["fallbackReason"] = fallbackReason
	}
	return marshalJSONString(payload)
}

func marshalJSONString(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(data)
}
