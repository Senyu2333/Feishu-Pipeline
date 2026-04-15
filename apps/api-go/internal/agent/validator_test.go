package agent

import "testing"

func TestParseStructuredJSON_StripsMarkdownFence(t *testing.T) {
	raw := "```json\n{\"title\":\"需求A\",\"summary\":\"摘要\",\"delivery_summary\":\"总结\",\"referenced_knowledge_titles\":[\"接口规范\"]}\n```"

	parsed, err := parseStructuredJSON[normalizedRequirement](raw)
	if err != nil {
		t.Fatalf("parseStructuredJSON returned error: %v", err)
	}

	if parsed.Title != "需求A" {
		t.Fatalf("unexpected title: %s", parsed.Title)
	}
	if len(parsed.ReferencedKnowledgeTitles) != 1 || parsed.ReferencedKnowledgeTitles[0] != "接口规范" {
		t.Fatalf("unexpected referenced knowledge titles: %#v", parsed.ReferencedKnowledgeTitles)
	}
}

func TestValidateTaskSplit_FillsDefaults(t *testing.T) {
	output := taskSplitOutput{
		Tasks: []taskPlan{
			{
				Type:         "frontend",
				Title:        "前端任务",
				Description:  "完成页面开发",
				EstimateDays: 0,
			},
		},
	}

	tasks, err := validateTaskSplit(output)
	if err != nil {
		t.Fatalf("validateTaskSplit returned error: %v", err)
	}

	if tasks[0].EstimateDays != 1 {
		t.Fatalf("unexpected estimate days: %d", tasks[0].EstimateDays)
	}
	if len(tasks[0].AcceptanceCriteria) == 0 {
		t.Fatal("acceptance criteria should be defaulted")
	}
	if len(tasks[0].Risks) == 0 {
		t.Fatal("risks should be defaulted")
	}
}
