package service

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"net/http"
	"net/url"
	"strings"
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

func (s *ChatService) RequestWithProxy(content, proxy string) (openai.ChatCompletionResponse, error) {
	proxyUrl, err := parseProxy(proxy)
	if err != nil {
		return openai.ChatCompletionResponse{}, err
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyUrl),
	}

	client := &http.Client{
		Transport: transport,
	}

	config := openai.DefaultConfig(s.token)
	config.HTTPClient = client

	openaiClient := openai.NewClientWithConfig(config)

	return openaiClient.CreateChatCompletion(
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
