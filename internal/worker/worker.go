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
var TooManyRequestsErr = fmt.Errorf("too many requests")

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

func (w *Worker) ConnectWithGeneratedProxy(connector service.ProxyConnecter, connectUrl string, proxy cfg.Proxy) error {
	extensionPath, err := w.GenerateProxyAuthExtension(proxy)
	if err != nil {
		// TODO: Подумать над возвращением ошибки
		log.Println("Generate proxy auth extension error:", err)
	}

	err = connector.ConnectWithProxy(connectUrl, extensionPath)
	if err != nil {
		return fmt.Errorf("selenium connect with proxy error:%w", err)
	}

	return nil
}

// Run должен быть вызван только после инициализации всех сервисов
func (w *Worker) Run() error {
	// TODO: подумать над релевантностью
	// Chat api test
	//if err := w.services.Chat.TestConnection(); err != nil {
	//	return fmt.Errorf("chat api connection error:%w", err)
	//}

	// Selenium parse page
	if err := w.services.Parse(w.d.BaseURL); err != nil {
		return fmt.Errorf("page parse error:%w", err)
	}
	log.Println("Web page parsed")

	// Page test
	if err := w.services.Selenium.TestPage(); err != nil {
		return TooManyRequestsErr
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
		return fmt.Errorf("go to visa type verification page error:%w", err)
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

		if err := w.services.Selenium.Wd().Get("https://russia.blsspainglobal.com/Global/Bls/VisaTypeVerification"); err != nil {
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
	err := w.RetryProcessCaptcha(w.d.CaptchaMaxTries)
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
	log.Println("Availability notification sent")

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
	if err := json.Unmarshal(cookiesJson, &cookies); err != nil {
		return fmt.Errorf("cannot unmarshal cookies:%w", err)
	}

	if err := w.services.Wd().DeleteAllCookies(); err != nil {
		return fmt.Errorf("cannot delete all cookies:%w", err)
	}

	if err := w.services.Selenium.SetCookies(cookies); err != nil {
		return fmt.Errorf("cannot set cookies:%w", err)
	}

	return w.services.Selenium.Refresh()
}

// SaveCookies сохраняет куки в файл. Функция вызывается в defer
func (w *Worker) SaveCookies() {
	var cookies []selenium.Cookie

	// TODO: refactor
	cookie, err := w.services.Wd().GetCookie(".AspNetCore.Cookies")
	if err != nil {
		log.Println("Cannot get cookies:%w", err)
		return
	}

	cookies = append(cookies, cookie)

	cookiesJson, err := json.Marshal(cookies)
	if err != nil {
		log.Println("Cannot marshal cookies:%w", err)
		return
	}

	if err := util.WriteFile(w.cookieFilePath(), cookiesJson); err != nil {
		log.Println("Cannot save cookies:%w", err)
		return
	}

	log.Println("Cookies saved")
}

func (w *Worker) cookieFilePath() string {
	return w.d.TmpFolder + w.d.CookieFile
}

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
