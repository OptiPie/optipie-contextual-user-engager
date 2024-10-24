package openaiapi

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sashabaranov/go-openai"
)

type NewOpenaiAPIArgs struct {
	OpenaiSecretKey string
	SystemMessage   string
}

type OpenaiAPI struct {
	client        *openai.Client
	systemMessage string
}

func NewOpenaiAPI(args NewOpenaiAPIArgs) (*OpenaiAPI, error) {
	if args.OpenaiSecretKey == "" {
		return nil, fmt.Errorf("openaiSecretKey can't be nil")
	}
	if args.SystemMessage == "" {
		return nil, fmt.Errorf("systemMessage can't be nil")
	}

	openaiClient := openai.NewClient(args.OpenaiSecretKey)
	return &OpenaiAPI{
		client:        openaiClient,
		systemMessage: args.SystemMessage,
	}, nil
}

func (oa *OpenaiAPI) CreateChat(ctx context.Context, message string) (string, error) {
	resp, err := oa.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: oa.systemMessage,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: message,
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	var content *ChatCompletionContent
	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &content)

	if err != nil {
		return "", fmt.Errorf("json unmarshall error %v", err)
	}

	if !content.IsRelated {
		return "", nil
	}

	return content.Reply, nil
}
