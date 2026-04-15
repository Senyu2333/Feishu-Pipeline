package agent

type normalizedRequirement struct {
	Title                     string   `json:"title"`
	Summary                   string   `json:"summary"`
	DeliverySummary           string   `json:"delivery_summary"`
	ReferencedKnowledgeTitles []string `json:"referenced_knowledge_titles"`
}

type taskSplitOutput struct {
	Tasks []taskPlan `json:"tasks"`
}

type taskPlan struct {
	Type               string   `json:"type"`
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
	Risks              []string `json:"risks"`
	Priority           string   `json:"priority"`
	EstimateDays       int      `json:"estimate_days"`
	AssigneeRole       string   `json:"assignee_role"`
}

type notificationOutput struct {
	Items []notificationPlan `json:"items"`
}

type notificationPlan struct {
	TaskTitle     string `json:"task_title"`
	NotifyContent string `json:"notify_content"`
}
