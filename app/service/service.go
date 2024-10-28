package service

import (
	"github.com/sashabaranov/go-openai"
	"github.com/tebeka/selenium"
)

type Selenium interface {
	Connect(url string) error
	ConnectWithProxy(url, extansionPath string) error
	GetCookies() ([]selenium.Cookie, error)
	SetCookies(cookies []selenium.Cookie) error
	DeleteCookie(key string) error
	Parse(url string) error
	MaximizeWindow() error

	GoTo(url string) error
	Refresh() error

	Wd() selenium.WebDriver
	TestPage() error

	IsAuthorized(neededURLPath string) (bool, error)

	ClickVerifyBtn() error

	PullCaptchaImage() ([]byte, error)
	SolveCaptcha(numbers []int) error
	Authorize() error
	BookNew() error
	BookNewAppointment() error
	CheckAvailability() (bool, error)

	Quit()
}

type Chat interface {
	TestConnection() error
	GetRespMsg(resp openai.ChatCompletionResponse) string
	Request3DOT5Turbo(content string) (openai.ChatCompletionResponse, error)
	Request4VPreviewWithImage(content, imageUrl string) (openai.ChatCompletionResponse, error)
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
	BaseURL string

	MaxTries int

	BlsEmail    string
	BlsPassword string

	ChatApiKey string

	ImgurClientId     string
	ImgurClientSecret string
}

func NewService(deps Deps) *Service {
	return &Service{
		Selenium: NewSeleniumService(deps.MaxTries, deps.BlsEmail, deps.BlsPassword),
		Chat:     NewChatService(deps.ChatApiKey),
		Image:    NewImageService(deps.ImgurClientId, deps.ImgurClientSecret),
	}
}
