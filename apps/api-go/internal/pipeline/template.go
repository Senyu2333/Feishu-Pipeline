package pipeline

import "encoding/json"

const DefaultTemplateID = "feature-delivery"

func DefaultTemplateDefinitionJSON() string {
	definition := map[string]any{
		"id":     DefaultTemplateID,
		"name":   "Feature Delivery",
		"stages": DefaultStageDefinitions,
	}
	bytes, err := json.Marshal(definition)
	if err != nil {
		return `{"id":"feature-delivery","name":"Feature Delivery"}`
	}
	return string(bytes)
}
