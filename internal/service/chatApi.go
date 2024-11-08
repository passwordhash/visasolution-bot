package service

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"net/http"
	cfg "visasolution/internal/config"
	pkgService "visasolution/pkg/service"
)

const testMsgReq = "Hello, World!"

type ChatService struct {
	token  string
	client *openai.Client
}

func NewChatService(token string) *ChatService {
	return &ChatService{token: token}
}

func (s *ChatService) ClientInitWithProxy(proxy cfg.Proxy) error {
	transport, err := pkgService.ProxyTransport(proxy.URL())
	if err != nil {
		return fmt.Errorf("failed to create proxy transport: %w", err)
	}

	client := &http.Client{
		Transport: transport,
	}

	config := openai.DefaultConfig(s.token)
	config.HTTPClient = client

	s.client = openai.NewClientWithConfig(config)

	return nil
}

func (s *ChatService) TestConnection() error {
	_, err := s.Request3DOT5Turbo(testMsgReq)
	return err
}

func (s *ChatService) Request3DOT5Turbo(content string) (openai.ChatCompletionResponse, error) {
	return s.client.CreateChatCompletion(
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
}

func (s *ChatService) Request4VPreviewWithImage(content, imageUrl string) (openai.ChatCompletionResponse, error) {
	return s.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleUser,
					MultiContent: []openai.ChatMessagePart{
						{
							Type: openai.ChatMessagePartTypeText,
							Text: content,
						},
						{
							Type: openai.ChatMessagePartTypeImageURL,
							ImageURL: &openai.ChatMessageImageURL{
								URL: imageUrl,
							},
						},
					},
				},
			},
		},
	)
}

func (s *ChatService) GetRespMsg(resp openai.ChatCompletionResponse) string {
	return resp.Choices[0].Message.Content
}
