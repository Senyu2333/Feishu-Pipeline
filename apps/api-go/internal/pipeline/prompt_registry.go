package pipeline

import (
	"encoding/json"
	"fmt"
	"strings"

	"feishu-pipeline/apps/api-go/internal/model"
)

const (
	AgentRequirementAnalyst = "requirement_analyst"
	AgentSolutionDesigner   = "solution_designer"
	AgentCodeGenerator      = "code_generator"
	AgentTestGenerator      = "test_generator"
	AgentCodeReviewer       = "code_reviewer"
	AgentDeliveryIntegrator = "delivery_integrator"
)

type AgentPromptSpec struct {
	StageKey       string
	AgentKey       string
	SystemPrompt   string
	UserPrompt     string
	RequiredFields []string
	ArtifactType   model.ArtifactType
	ArtifactTitle  string
}

type StagePromptDefinition struct {
	StageKey       string
	AgentKey       string
	Role           string
	OutputContract string
	RequiredFields []string
	ArtifactType   model.ArtifactType
	ArtifactTitle  string
}

type PromptRegistry struct {
	definitions map[string]StagePromptDefinition
}

func DefaultPromptRegistry() *PromptRegistry {
	definitions := []StagePromptDefinition{
		{
			StageKey:       StageRequirementAnalysis,
			AgentKey:       AgentRequirementAnalyst,
			Role:           "你是需求分析 Agent，负责把自然语言需求转成稳定、可审批、可被后续阶段消费的结构化需求。",
			OutputContract: `{"summary":"","goals":[],"nonGoals":[],"acceptanceCriteria":[],"questions":[],"risks":[]}`,
			RequiredFields: []string{SchemaFieldSummary, SchemaFieldGoals, SchemaFieldAcceptanceCriteria},
			ArtifactType:   model.ArtifactStructuredRequirement,
			ArtifactTitle:  "结构化需求",
		},
		{
			StageKey:       StageSolutionDesign,
			AgentKey:       AgentSolutionDesigner,
			Role:           "你是方案设计 Agent，负责结合需求、仓库上下文和历史审批意见，输出技术方案和影响范围。",
			OutputContract: `{"summary":"","impactFiles":[],"apiChanges":[],"dataModelChanges":[],"implementationPlan":[],"risks":[]}`,
			RequiredFields: []string{SchemaFieldSummary, SchemaFieldImpactFiles, SchemaFieldImplementationPlan},
			ArtifactType:   model.ArtifactSolutionDesign,
			ArtifactTitle:  "技术方案",
		},
		{
			StageKey:       StageCodeGeneration,
			AgentKey:       AgentCodeGenerator,
			Role:           "你是代码生成 Agent，负责输出受控代码变更计划。当前阶段只能生成结构化 patch 计划，不允许要求直接写文件、删除文件或推送远程仓库。",
			OutputContract: `{"summary":"","changedFiles":[],"patches":[],"diffSummary":[],"manualSteps":[]}`,
			RequiredFields: []string{SchemaFieldSummary, SchemaFieldChangedFiles, SchemaFieldDiffSummary},
			ArtifactType:   model.ArtifactCodeDiff,
			ArtifactTitle:  "代码变更计划",
		},
		{
			StageKey:       StageTestGeneration,
			AgentKey:       AgentTestGenerator,
			Role:           "你是测试生成 Agent，负责根据需求和代码变更计划输出测试计划。命令执行由后端白名单控制，你只输出建议和结构化结果。",
			OutputContract: `{"summary":"","testPlan":[],"commands":[],"commandResults":[],"status":""}`,
			RequiredFields: []string{SchemaFieldSummary, SchemaFieldTestPlan, SchemaFieldStatus},
			ArtifactType:   model.ArtifactTestReport,
			ArtifactTitle:  "测试报告",
		},
		{
			StageKey:       StageCodeReview,
			AgentKey:       AgentCodeReviewer,
			Role:           "你是代码评审 Agent，负责从正确性、安全性、可维护性、测试充分性维度审查变更。",
			OutputContract: `{"summary":"","conclusion":"","issues":[],"securityNotes":[],"maintainabilityNotes":[],"testCoverageNotes":[]}`,
			RequiredFields: []string{SchemaFieldSummary, SchemaFieldConclusion, SchemaFieldIssues},
			ArtifactType:   model.ArtifactReviewReport,
			ArtifactTitle:  "评审报告",
		},
		{
			StageKey:       StageDelivery,
			AgentKey:       AgentDeliveryIntegrator,
			Role:           "你是交付集成 Agent，负责汇总需求、方案、测试和评审产物，生成安全的 PR/MR 草稿和发布说明。当前阶段不执行 push 或创建远程 PR/MR。",
			OutputContract: `{"summary":"","changedFiles":[],"validation":[],"prTitle":"","prBody":"","manualReleaseNotes":[]}`,
			RequiredFields: []string{SchemaFieldSummary, SchemaFieldPRTitle, SchemaFieldPRBody},
			ArtifactType:   model.ArtifactDeliverySummary,
			ArtifactTitle:  "交付摘要",
		},
	}
	registry := &PromptRegistry{definitions: map[string]StagePromptDefinition{}}
	for _, definition := range definitions {
		registry.definitions[definition.StageKey] = definition
	}
	return registry
}

func (r *PromptRegistry) Build(stageContext StageContext) (AgentPromptSpec, bool) {
	if r == nil {
		return AgentPromptSpec{}, false
	}
	definition, ok := r.definitions[stageContext.Stage.StageKey]
	if !ok {
		return AgentPromptSpec{}, false
	}
	inputJSON, _ := json.MarshalIndent(stageContext.Input, "", "  ")
	systemPrompt := strings.Join([]string{
		definition.Role,
		"你在 DevFlow Engine 的一个 Pipeline Stage 中工作。",
		"你必须只输出合法 JSON，不要输出 Markdown 代码块、解释、前后缀说明。",
		"输出 JSON 必须满足以下结构，字段名必须保持一致：",
		definition.OutputContract,
		"如果信息不足，在 risks 或 questions 字段中说明，不要编造不存在的文件或外部状态。",
	}, "\n")
	userPrompt := fmt.Sprintf("当前 Pipeline 阶段：%s\nRun 标题：%s\n目标仓库：%s\n目标分支：%s\n工作分支：%s\n\n阶段输入 JSON：\n%s\n\n请按系统要求输出本阶段 JSON。",
		stageContext.Stage.StageKey,
		stageContext.Run.Title,
		stageContext.Run.TargetRepo,
		stageContext.Run.TargetBranch,
		stageContext.Run.WorkBranch,
		string(inputJSON),
	)
	return AgentPromptSpec{
		StageKey:       definition.StageKey,
		AgentKey:       definition.AgentKey,
		SystemPrompt:   systemPrompt,
		UserPrompt:     userPrompt,
		RequiredFields: definition.RequiredFields,
		ArtifactType:   definition.ArtifactType,
		ArtifactTitle:  definition.ArtifactTitle,
	}, true
}
