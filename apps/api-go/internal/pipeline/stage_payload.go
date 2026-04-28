package pipeline

import "feishu-pipeline/apps/api-go/internal/model"

func impactFilesFromContext(context RepositoryContext) []string {
	result := make([]string, 0, len(context.CandidateFiles))
	seen := map[string]bool{}
	for _, item := range context.CandidateFiles {
		if seen[item.Path] {
			continue
		}
		seen[item.Path] = true
		result = append(result, item.Path)
	}
	return result
}

func nestedString(input map[string]any, key string, artifactType model.ArtifactType, field string) string {
	latestArtifacts, ok := input[key].(map[string]any)
	if !ok {
		return ""
	}
	artifact, ok := latestArtifacts[string(artifactType)].(map[string]any)
	if !ok {
		return ""
	}
	value, _ := artifact[field].(string)
	return value
}

func nestedStringSlice(input map[string]any, key string, artifactType model.ArtifactType, field string) []string {
	latestArtifacts, ok := input[key].(map[string]any)
	if !ok {
		return nil
	}
	artifact, ok := latestArtifacts[string(artifactType)].(map[string]any)
	if !ok {
		return nil
	}
	items, ok := artifact[field].([]any)
	if !ok {
		return nil
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		text, ok := item.(string)
		if ok && text != "" {
			result = append(result, text)
		}
	}
	return result
}

func nestedMapSlice(input map[string]any, key string, artifactType model.ArtifactType, field string) []map[string]any {
	latestArtifacts, ok := input[key].(map[string]any)
	if !ok {
		return nil
	}
	artifact, ok := latestArtifacts[string(artifactType)].(map[string]any)
	if !ok {
		return nil
	}
	items, ok := artifact[field].([]any)
	if !ok {
		return nil
	}
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		entry, ok := item.(map[string]any)
		if ok {
			result = append(result, entry)
		}
	}
	return result
}

func checkpointSummaries(items []model.Checkpoint) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, map[string]any{"id": item.ID, "status": item.Status, "decision": item.Decision, "comment": item.Comment})
	}
	return result
}

func artifactTitles(items []model.Artifact) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		if item.Title != "" {
			result = append(result, item.Title)
		}
	}
	return result
}
