package service

import (
	"github.com/sashabaranov/go-openai"
	"github.com/tebeka/selenium"
	cfg "visasolution/internal/config"
)

type ProxyConnecter interface {
	// TODO
	ConnectWithProxy(extansionPath string) error
}

type Proxier interface {
	ClientInitWithProxy(proxy cfg.Proxy) error
}

type Selenium interface {
	ConnectWithProxy(extansionPath string) error
	AuthCookie() (selenium.Cookie, error)
	Cookies() ([]selenium.Cookie, error)
	SetCookies(cookies []selenium.Cookie) error
	DeleteCookie(key string) error
	DeleteAllCookies() error
	MaximizeWindow() error

	GoTo(url string) error
	Refresh() error

	TestPage() error

	IsAuthorized(neededURLPath string) (bool, error)

	ClickVerifyBtn() error

	PullPageScreenshot() ([]byte, error)
	PullCaptchaImage() ([]byte, error)
	SolveCaptcha(numbers []int) error
	Authorize() error
	BookNew() error
	BookNewAppointment() error
	CheckAvailability() (bool, error)

	Quit() error
}

type Chat interface {
	TestConnection() error
	GetRespMsg(resp openai.ChatCompletionResponse) string
	Request3DOT5Turbo(content string) (openai.ChatCompletionResponse, error)
	Request4VPreviewWithImage(content, imageUrl string) (openai.ChatCompletionResponse, error)
	Proxier
}

type Image interface {
	Proxier
	UploadImage(imagePath string) (string, error)
}

type Email interface {
	SendAvailbilityNotification(to string) error
}

type Service struct {
	Selenium
	Chat
	Image
	Email
}

type Deps struct {
	SeleniumURL string
	BaseURL     string

	MaxTries int

	BlsEmail    string
	BlsPassword string

	ChatApiKey string

	ImgurClientId     string
	ImgurClientSecret string

	EmailDeps
}

func NewService(deps Deps) *Service {
	return &Service{
		Selenium: NewSeleniumService(deps.MaxTries, deps.BlsEmail, deps.SeleniumURL, deps.BlsPassword),
		Chat:     NewChatService(deps.ChatApiKey),
		Image:    NewImageService(deps.ImgurClientId, deps.ImgurClientSecret),
		Email:    NewEmailService(deps.EmailDeps),
	}
}
