package worker

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tebeka/selenium"
	"log"
	"os"
	"time"
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

	NotifiedEmail   string
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
	if err := util.CreateFolder(w.d.TmpFolder); err != nil {
		return fmt.Errorf("cannot create folder:%w", err)
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
// TODO: refactor
func (w *Worker) Run() error {
	err := w.services.Selenium.GoTo(w.d.BaseURL)
	if errors.Is(err, service.InvalidSessionError) {
		return WDConnectError{Msg: err.Error()}
	}
	if err != nil {
		return fmt.Errorf("page parse error:%w", err)
	}
	log.Println("Web page parsed")

	// Page test
	if err := w.services.Selenium.TestPage(); err != nil {
		return TooManyRequestsErr{Msg: "test page error"}
	}
	log.Println("Page successfully loaded")

	// Maximize window
	if err := w.services.Selenium.MaximizeWindow(); err != nil {
		return fmt.Errorf("cannot maximize window:%w", err)
	}

	// Load cookies
	if err := w.LoadCookies(); err != nil {
		log.Println("Cookies load error:", err)
	}

	// Go to visa type verification page
	if err := w.services.Selenium.GoTo(w.d.BaseURL + w.d.VisaTypeURL); err != nil {
		//return fmt.Errorf("go to visa type verification page error:%w", err)
		return TooManyRequestsErr{Msg: "go to visa type verification page error"}
	}

	isAuthorized, _ := w.services.Selenium.IsAuthorized(w.d.BaseURL + w.d.VisaTypeURL)
	if !isAuthorized {
		// Solving first captcha
		if err := w.services.Selenium.ClickVerifyBtn(); err != nil {
			return fmt.Errorf("click verify first catpcha error:%w", err)
		}

		log.Println("Retry process first captcha starts ...")
		err := w.RetryProcessCaptcha(w.d.CaptchaMaxTries)
		if errors.Is(err, service.InvalidSelectionError) {
			return service.InvalidSelectionError
		}
		if err != nil {
			return fmt.Errorf("retry process first captcha error:%w", err)
		}
		log.Println("Retry process first captcha successfully ended")

		// TODO: сделать ожидание прогрузки
		time.Sleep(time.Second * 3)

		// Authorization
		if err := w.services.Selenium.Authorize(); err != nil {
			return fmt.Errorf("authorization error:%w", err)
		}
		log.Println("Authorization successfully")

		// TODO: реализовать ожидание подгрузки следующий страницы (ожидание момента авторизации) >>>
		time.Sleep(time.Second * 5)

		err = w.services.Selenium.GoTo(w.d.BaseURL + w.d.VisaTypeURL)
		if err != nil {
			return err
		}

		time.Sleep(time.Second * 5)
		// TODO: <<<

		w.SaveCookies()
	} else {
		log.Println("Already authorized. Skip authorization")
	}

	// Solving second captcha
	if err := w.services.Selenium.ClickVerifyBtn(); err != nil {
		return fmt.Errorf("click verify second captcha error:%w", err)
	}
	log.Println("Second captcha successfully clicked")

	// DEBUG:
	time.Sleep(time.Second * 3)

	log.Println("Retry process second captcha starts ...")
	err = w.RetryProcessCaptcha(w.d.CaptchaMaxTries)
	if errors.Is(err, service.InvalidSelectionError) {
		return service.InvalidSelectionError
	}
	if err != nil {
		return fmt.Errorf("retry process first captcha error:%w", err)
	}
	log.Println("Retry process second captcha successfully ended")

	// Book new appointment
	if err := w.services.Selenium.BookNewAppointment(); err != nil {
		return fmt.Errorf("book new appointment error:%w", err)
	}
	log.Println("Book new appointment submit successfully")

	// DEBUG:
	time.Sleep(time.Second * 4)

	// Check availability
	isAppointmentAvailable, err := w.services.Selenium.CheckAvailability()
	if err != nil {
		return fmt.Errorf("check availability error:%w", err)
	}
	log.Println("Check availability successfully")

	if isAppointmentAvailable {
		log.Println("!!!Appointment available!!!")

	} else {
		log.Println("!!!Appointment NOT available!!!")
	}

	// TEMP: Save page screenshot
	if err := w.savePageScreenshot(); err != nil {
		log.Println("Cannot save page screenshot:%w", err)
	}

	// TEMP: в качестве проверки правильности определния мест для записи, в любом случае отправляется письм
	err = w.services.Email.SendAvailbilityNotification(w.d.NotifiedEmail)
	if err != nil {
		return fmt.Errorf("send availability notification error:%w", err)
	}
	log.Println("Availability notification sent to ", w.d.NotifiedEmail)

	// DEBUG:
	time.Sleep(time.Second * 15)

	log.Println("Work done")

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

	if err := util.WriteFile(path, data); err != nil {
		return fmt.Errorf("cannot write screenshot:%w", err)
	}

	return nil
}

func (w *Worker) cookieFilePath() string {
	return w.d.TmpFolder + w.d.CookieFile
}

func LoadProxies(filePath string) (*cfg.ProxiesManager, error) {
	proxiesFile, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read proxies file: %v", err)
	}
	return cfg.ParseProxiesFile(proxiesFile)
}
