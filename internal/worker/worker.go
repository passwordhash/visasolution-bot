package worker

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tebeka/selenium"
	"log"
	"os"
	cfg "visasolution/internal/config"
	"visasolution/internal/service"
	"visasolution/pkg/util"
)

// TooManyRequestsErr ошибка, возникающая при превышении лимита запросов к ресурсу
type TooManyRequestsErr struct {
	Msg string
}

func (e TooManyRequestsErr) Error() string {
	return fmt.Sprintf("too many requests error: %s", e.Msg)
}

// WDConnectError ошибка подключения к Selenium WebDriver
type WDConnectError struct {
	Msg string
}

func (e WDConnectError) Error() string {
	return fmt.Sprintf("webdriver connect error: %s", e.Msg)
}

type Deps struct {
	BaseURL     string
	VisaTypeURL string

	TmpFolder      string
	CookieFile     string
	ScreenshotFile string

	NotifiedEmails  []string
	CaptchaMaxTries int
}

type Worker struct {
	services *service.Service
	d        Deps
}

func NewWorker(services *service.Service, emailDeps Deps) *Worker {
	return &Worker{
		services: services,
		d:        emailDeps,
	}
}

// MakePreparation выполняет подготовительную работу
func (w *Worker) MakePreparation() error {
	err := util.CreateFolder(w.d.TmpFolder)
	if err != nil {
		return fmt.Errorf("cannot create tmp folder:%w", err)
	}

	return nil
}

// ConnectSameProxy выполняет подключение к Selenium WebDriver с текущим прокси
func (w *Worker) ConnectSameProxy(connector service.ProxyConnecter) error {
	err := connector.ConnectWithProxy(w.chromeExtensionPath())
	if err != nil {
		return fmt.Errorf("selenium connect with proxy error:%w", err)
	}

	return nil
}

// ConnectGeneratedProxy выполняет подключение к Selenium WebDriver с новым прокси
// Функция генерирует расширение для авторизации прокси
func (w *Worker) ConnectGeneratedProxy(connector service.ProxyConnecter, proxy cfg.Proxy) error {
	extensionPath, err := w.GenerateProxyAuthExtension(proxy)
	if err != nil {
		// TODO: Подумать над возвращением ошибки
		log.Println("Generate proxy auth extension error:", err)
	}

	err = connector.ConnectWithProxy(extensionPath)
	if err != nil {
		return fmt.Errorf("selenium connect with new generated proxy error:%w", err)
	}

	return nil
}

// Run должен быть вызван только после инициализации всех сервисов.
// Функция выполняет основной алгоритм работы бота.
func (w *Worker) Run() error {
	err := w.services.Selenium.GoTo(w.d.BaseURL)
	if errors.Is(err, service.InvalidSessionError) {
		return WDConnectError{Msg: err.Error()}
	}
	if err != nil {
		return fmt.Errorf("page parse error:%w", err)
	}
	log.Println("Web page parsed")

	err = w.services.Selenium.MaximizeWindow()
	if err != nil {
		return fmt.Errorf("cannot maximize window:%w", err)
	}

	err = w.LoadCookies()
	if err != nil {
		log.Println("Cookies load error:", err)
	}

	err = w.services.Selenium.GoTo(w.d.BaseURL + w.d.VisaTypeURL)
	if err != nil {
		return TooManyRequestsErr{Msg: "go to visa type verification page error"}
	}

	isAuthorized, _ := w.services.Selenium.IsAuthorized(w.d.BaseURL + w.d.VisaTypeURL)
	if !isAuthorized {
		err := w.handleAuthorization()
		if err != nil {
			return fmt.Errorf("authorization error:%w", err)
		}
	} else {
		log.Println("Already authorized. Skip authorization")
	}

	// Solving second captcha
	err = w.handleCaptcha()
	if err != nil {
		return fmt.Errorf("second captcha error:%w", err)
	}

	err = w.services.Selenium.BookNewAppointment()
	if err != nil {
		return fmt.Errorf("book new appointment error:%w", err)
	}
	log.Println("Book new appointment submit successfully")

	isAppointmentAvailable, err := w.services.Selenium.CheckAvailability()
	if err != nil {
		return fmt.Errorf("check availability error:%w", err)
	}

	if isAppointmentAvailable {
		log.Println("!!!Appointment available!!!")
	} else {
		log.Println("!!!Appointment NOT available!!!")
	}

	// TEMP: Save page screenshot
	err = w.savePageScreenshot()
	if err != nil {
		log.Println("Cannot save page screenshot:%w", err)
	}

	// TEMP: в качестве проверки правильности определния мест для записи, в любом случае отправляется письм
	sent, err := w.services.Email.SendBulkNotification(w.d.NotifiedEmails)
	w.handleEmailsSent(sent, err)

	log.Println("Work done")

	return nil
}

// handleAuthorization обрабатывает авторизацию на сайте.
// Функция вызывается в случае, если необходимо авторизоваться.
func (w *Worker) handleAuthorization() error {
	err := w.handleCaptcha()
	if err != nil {
		return fmt.Errorf("authorization captcha error:%w", err)
	}

	log.Println("Retry process first captcha successfully ended")

	err = w.services.Selenium.Authorize()
	if err != nil {
		return fmt.Errorf("authorization error:%w", err)
	}

	log.Println("Authorization successfully ended")

	err = w.services.Selenium.GoTo(w.d.BaseURL + w.d.VisaTypeURL)
	if err != nil {
		return err
	}

	w.SaveCookies()

	return nil
}

// handleCaptcha выполняет обработку имеющейся на странице капчи
func (w *Worker) handleCaptcha() error {
	err := w.services.Selenium.ClickVerifyBtn()
	if err != nil {
		return fmt.Errorf("click verify captcha error:%w", err)
	}

	log.Println("Retry process captcha starts ...")

	err = w.RetryProcessCaptcha(w.d.CaptchaMaxTries)
	if errors.Is(err, service.InvalidSelectionError) {
		return service.InvalidSelectionError
	}
	if err != nil {
		return fmt.Errorf("retry process captcha error:%w", err)
	}

	return nil
}

func (w *Worker) LoadCookies() error {
	cookiesJson, err := os.ReadFile(w.cookieFilePath())
	if err != nil {
		return fmt.Errorf("cannot read cookies:%w", err)
	}

	var cookies []selenium.Cookie
	err = json.Unmarshal(cookiesJson, &cookies)
	if err != nil {
		return fmt.Errorf("cannot unmarshal cookies:%w", err)
	}

	err = w.services.Selenium.DeleteAllCookies()
	if err != nil {
		return fmt.Errorf("cannot delete all cookies:%w", err)
	}

	err = w.services.Selenium.SetCookies(cookies)
	if err != nil {
		return fmt.Errorf("cannot set cookies:%w", err)
	}

	return w.services.Selenium.Refresh()
}

// SaveCookies сохраняет куки в файл. Функция вызывается в defer
func (w *Worker) SaveCookies() {
	var cookies []selenium.Cookie

	cookie, err := w.services.Selenium.AuthCookie()
	if err != nil {
		log.Println("cannot get auth cookie: ", err)
		return
	}

	cookies = append(cookies, cookie)

	cookiesJson, err := json.Marshal(cookies)
	if err != nil {
		log.Println("Cannot marshal cookies:%w", err)
		return
	}

	err = util.WriteFile(w.cookieFilePath(), cookiesJson)
	if err != nil {
		log.Println("Cannot save cookies:%w", err)
		return
	}

	log.Println("Cookies saved")
}

// savePageScreenshot сохраняет скриншот страницы.
// Временная функция для отладки
func (w *Worker) savePageScreenshot() error {
	path := w.d.TmpFolder + w.d.ScreenshotFile
	data, err := w.services.Selenium.PullPageScreenshot()
	if err != nil {
		return fmt.Errorf("cannot pull page screenshot:%w", err)
	}

	err = util.WriteFile(path, data)
	if err != nil {
		return fmt.Errorf("cannot write screenshot:%w", err)
	}

	return nil
}

func (w *Worker) cookieFilePath() string {
	return w.d.TmpFolder + w.d.CookieFile
}

func (w *Worker) handleEmailsSent(sent []string, err error) {
	var sendErr service.SendEmailsError
	if errors.Is(err, service.SendEmailsError{}) {
		log.Println(sendErr)
	}
	if err != nil {
		log.Println("Error sending email:", err)
	}

	log.Println("Emails sent:", sent)
}

// LoadProxies загружает прокси из файла
func LoadProxies(filePath string) (*cfg.ProxiesManager, error) {
	proxiesFile, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read proxies file: %v", err)
	}

	return cfg.ParseProxiesFile(proxiesFile)
}
