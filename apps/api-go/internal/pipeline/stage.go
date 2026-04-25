package pipeline

import "feishu-pipeline/apps/api-go/internal/model"

type StageDefinition struct {
	Key          string
	Name         string
	Type         model.StageType
	Order        int
	IsCheckpoint bool
}

const (
	StageRequirementAnalysis = "requirement_analysis"
	StageSolutionDesign      = "solution_design"
	StageCheckpointDesign    = "checkpoint_design"
	StageCodeGeneration      = "code_generation"
	StageTestGeneration      = "test_generation"
	StageCodeReview          = "code_review"
	StageCheckpointReview    = "checkpoint_review"
	StageDelivery            = "delivery"
)

var DefaultStageDefinitions = []StageDefinition{
	{Key: StageRequirementAnalysis, Name: "需求分析", Type: model.StageTypeAnalysis, Order: 1},
	{Key: StageSolutionDesign, Name: "方案设计", Type: model.StageTypeDesign, Order: 2},
	{Key: StageCheckpointDesign, Name: "方案审批", Type: model.StageTypeCheckpoint, Order: 3, IsCheckpoint: true},
	{Key: StageCodeGeneration, Name: "代码生成", Type: model.StageTypeCodegen, Order: 4},
	{Key: StageTestGeneration, Name: "测试生成", Type: model.StageTypeTest, Order: 5},
	{Key: StageCodeReview, Name: "代码评审", Type: model.StageTypeReview, Order: 6},
	{Key: StageCheckpointReview, Name: "评审确认", Type: model.StageTypeCheckpoint, Order: 7, IsCheckpoint: true},
	{Key: StageDelivery, Name: "交付集成", Type: model.StageTypeDelivery, Order: 8},
}
