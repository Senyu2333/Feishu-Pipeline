package pipeline

import "feishu-pipeline/apps/api-go/internal/model"

func CanStartRun(status model.PipelineRunStatus) bool {
	switch status {
	case model.PipelineRunDraft, model.PipelineRunFailed:
		return true
	default:
		return false
	}
}

func CanPauseRun(status model.PipelineRunStatus) bool {
	switch status {
	case model.PipelineRunQueued, model.PipelineRunRunning:
		return true
	default:
		return false
	}
}

func CanResumeRun(status model.PipelineRunStatus) bool {
	switch status {
	case model.PipelineRunPaused, model.PipelineRunFailed:
		return true
	default:
		return false
	}
}

func CanTerminateRun(status model.PipelineRunStatus) bool {
	switch status {
	case model.PipelineRunDraft, model.PipelineRunQueued, model.PipelineRunRunning, model.PipelineRunWaitingApproval, model.PipelineRunPaused, model.PipelineRunFailed:
		return true
	default:
		return false
	}
}

func CanApproveCheckpoint(run model.PipelineRun, stage model.StageRun, checkpoint model.Checkpoint) bool {
	return canDecideCheckpoint(run, stage, checkpoint)
}

func CanRejectCheckpoint(run model.PipelineRun, stage model.StageRun, checkpoint model.Checkpoint) bool {
	return canDecideCheckpoint(run, stage, checkpoint)
}

func canDecideCheckpoint(run model.PipelineRun, stage model.StageRun, checkpoint model.Checkpoint) bool {
	return run.Status == model.PipelineRunWaitingApproval && stage.Status == model.StageRunWaitingApproval && checkpoint.Status == model.CheckpointPending && checkpoint.PipelineRunID == run.ID && checkpoint.StageRunID == stage.ID
}
