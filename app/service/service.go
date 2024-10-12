package service

import (
	"github.com/sashabaranov/go-openai"
	"github.com/tebeka/selenium"
)

type Selenium interface {
	Parse(url string) error
	Wd() selenium.WebDriver
	MaximizeWindow() error
	ProcessCaptcha() error
	Connect(url string) error
	Quit()
}

type Chat interface {
	GetRespMsg(resp openai.ChatCompletionResponse) string
	Request3DOT5Turbo(content string) (openai.ChatCompletionResponse, error)
	ClientInitWithProxy(proxy string) error
}

type Image interface {
	UploadImage(url string) (string, error)
}

type Service struct {
	Selenium
	Chat
	Image
}

type Deps struct {
	MaxTries int

	BlsEmail    string
	BlsPassword string

	ChatApiKey string

	ImgurClientId     string
	ImgurClientSecret string
}

func NewService(deps Deps) *Service {
	return &Service{
		Selenium: NewSeleniumService(deps.MaxTries),
		Chat:     NewChatService(deps.ChatApiKey),
		Image:    NewImageService(deps.ImgurClientId, deps.ImgurClientSecret),
	}
}
