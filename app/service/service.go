package service

import (
	"github.com/sashabaranov/go-openai"
	"github.com/tebeka/selenium"
)

type Selenium interface {
	Wd() selenium.WebDriver
	ProcessCaptcha(wd selenium.WebDriver) error
	Connect(url string) (selenium.WebDriver, error)
	Quit()
}

type Chat interface {
	GPT3DOT5TurboRequest(content string) (openai.ChatCompletionResponse, error)
	GetRespMsg(resp openai.ChatCompletionResponse) string
}

type Service struct {
	Selenium
	Chat
}

type Deps struct {
	MaxTries int

	BlsEmail    string
	BlsPassword string

	ChatApiKey string
}

func NewService(deps Deps) *Service {
	return &Service{
		Selenium: NewSeleniumService(deps.MaxTries),
		Chat:     NewChatService(deps.ChatApiKey),
	}
}
