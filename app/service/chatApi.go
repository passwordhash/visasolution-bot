package service

import (
	"context"
	"github.com/sashabaranov/go-openai"
)

type ChatService struct {
	token string
}

func NewChatService(token string) *ChatService {
	return &ChatService{token: token}
}

func (s *ChatService) GetRespMsg(resp openai.ChatCompletionResponse) string {
	return resp.Choices[0].Message.Content
}

func (s *ChatService) GPT3DOT5TurboRequest(content string) (openai.ChatCompletionResponse, error) {
	client := openai.NewClient(s.token)
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
