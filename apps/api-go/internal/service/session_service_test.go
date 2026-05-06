package service

import "testing"

func TestIsPublishIntentRecognizesAssistantWorkflowConfirmation(t *testing.T) {
	cases := []string{
		"好的，系统将自动触发需求交付工作流，请等待后续流程推进。",
		"需求确认完成，进入流水线",
		"可以开始交付",
	}
	for _, item := range cases {
		if !isPublishIntent(item) {
			t.Fatalf("expected publish intent for %q", item)
		}
	}
}
