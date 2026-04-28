package pipeline

import "encoding/json"

const DefaultTemplateID = "feature-delivery"

type TemplateDefinition struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Stages []StageDefinition `json:"stages"`
}

func DefaultTemplateDefinitionJSON() string {
	definition := TemplateDefinition{
		ID:     DefaultTemplateID,
		Name:   "Feature Delivery",
		Stages: DefaultStageDefinitions,
	}
	bytes, err := json.Marshal(definition)
	if err != nil {
		return `{"id":"feature-delivery","name":"Feature Delivery"}`
	}
	return string(bytes)
}

func ParseTemplateDefinition(raw string) (TemplateDefinition, error) {
	if raw == "" {
		return TemplateDefinition{ID: DefaultTemplateID, Name: "Feature Delivery", Stages: DefaultStageDefinitions}, nil
	}
	var definition TemplateDefinition
	if err := json.Unmarshal([]byte(raw), &definition); err != nil {
		return TemplateDefinition{}, err
	}
	if len(definition.Stages) == 0 {
		definition.Stages = DefaultStageDefinitions
	}
	return definition, nil
}
