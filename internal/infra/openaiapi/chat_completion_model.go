package openaiapi

type ChatCompletionContent struct {
	IsRelated bool   `json:"isRelated"`
	Reply     string `json:"reply"`
}
