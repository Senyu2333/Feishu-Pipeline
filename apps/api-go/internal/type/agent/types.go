package agenttype

import "feishu-pipeline/apps/api-go/internal/model"

type SessionAggregate struct {
	Session      model.Session
	Owner        model.User
	MessageCount int
	Messages     []model.Message
	Requirement  *model.Requirement
	Tasks        []model.Task
}
