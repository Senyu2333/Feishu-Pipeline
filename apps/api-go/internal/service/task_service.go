package service

import (
	"context"

	"feishu-pipeline/apps/api-go/internal/external/feishu"
	"feishu-pipeline/apps/api-go/internal/model"
	"feishu-pipeline/apps/api-go/internal/repo"
)

type TaskService struct {
	repository   *repo.Repository
	feishuClient *feishu.Client
}

func NewTaskService(repository *repo.Repository, feishuClient *feishu.Client) *TaskService {
	return &TaskService{
		repository:   repository,
		feishuClient: feishuClient,
	}
}

func (s *TaskService) GetTask(ctx context.Context, taskID string) (model.Task, error) {
	return s.repository.GetTaskByID(ctx, taskID)
}

func (s *TaskService) UpdateTaskStatus(ctx context.Context, taskID string, status model.TaskStatus) (model.Task, error) {
	task, err := s.repository.UpdateTaskStatus(ctx, taskID, status)
	if err != nil {
		return model.Task{}, err
	}

	recordResult, err := s.feishuClient.UpsertTaskRecord(ctx, task)
	if err == nil && recordResult.RecordURL != "" {
		task.BitableAppToken = recordResult.AppToken
		task.BitableTableID = recordResult.TableID
		task.BitableRecordID = recordResult.RecordID
		task.BitableRecordURL = recordResult.RecordURL
		_ = s.repository.UpdateTaskLinks(ctx, task.ID, task.DocURL, recordResult.RecordURL, recordResult.RecordID, recordResult.AppToken, recordResult.TableID)
	}
	return task, nil
}
