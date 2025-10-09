package models

type ToolArgProps struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type ToolFuncParams struct {
	Type       string                  `json:"type"`
	Properties map[string]ToolArgProps `json:"properties"`
	Required   []string                `json:"required"`
}

type ToolFunc struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  ToolFuncParams `json:"parameters"`
}

type Tool struct {
	Type     string   `json:"type"`
	Function ToolFunc `json:"function"`
}
