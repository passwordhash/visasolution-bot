package api

import (
	"context"
	"github.com/sashabaranov/go-openai"
)

const TestMsgReq = "Если работаешь, отправь 'Hello, Worlrd!'"

func GetRespMsg(resp openai.ChatCompletionResponse) string {
	return resp.Choices[0].Message.Content
}

func GPT3DOT5TurboRequest(content, token string) (openai.ChatCompletionResponse, error) {
	client := openai.NewClient(token)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: content,
				},
			},
		},
	)

	return resp, err
}
