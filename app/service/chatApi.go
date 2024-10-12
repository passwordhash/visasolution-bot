package service

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"net/http"
	"net/url"
	"strings"
)

const testMsgReq = "Hello, World!"

type ChatService struct {
	token  string
	client *openai.Client
}

func NewChatService(token string) *ChatService {
	return &ChatService{token: token}
}

func (s *ChatService) TestConnection() error {
	_, err := s.Request3DOT5Turbo(testMsgReq)
	return err
}

func (s *ChatService) GetRespMsg(resp openai.ChatCompletionResponse) string {
	return resp.Choices[0].Message.Content
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

func (s *ChatService) ClientInitWithProxy(proxy string) error {
	proxyUrl, err := parseProxy(proxy)
	if err != nil {
		return err
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyUrl),
	}

	client := &http.Client{
		Transport: transport,
	}

	config := openai.DefaultConfig(s.token)
	config.HTTPClient = client

	s.client = openai.NewClientWithConfig(config)

	return nil
}

func (s *ChatService) RequestCaptchaWithProxy(content, proxy string) (openai.ChatCompletionResponse, error) {
	return s.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4VisionPreview,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: content,
				},
			},
		},
	)
}

func parseProxy(proxyString string) (*url.URL, error) {
	// Разделяем строку на части
	parts := strings.Split(proxyString, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid proxy format")
	}

	// Прокси и учетные данные
	proxyAddress := parts[0]
	credentials := parts[1]

	// Создаем URL-адрес прокси
	proxyURL, err := url.Parse(fmt.Sprintf("http://%s@%s", credentials, proxyAddress))
	if err != nil {
		return nil, err
	}

	return proxyURL, nil
}
