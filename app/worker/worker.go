package worker

import (
	"encoding/json"
	"fmt"
	"github.com/tebeka/selenium"
	"log"
	"os"
	"time"
	"visasolution/app/service"
	"visasolution/app/util"
)

const (
	tmpFolder  = "tmp/"
	cookieFile = "cookies.json"
)

var cookiePath = tmpFolder + cookieFile

// TODO: change
const availbilityNotifiedEmail = "iam@it-yaroslav.ru"

type Worker struct {
	services    *service.Service
	baseURL     string
	visaTypeURL string
}

func NewWorker(services *service.Service, parseUrl, visaTypeUrl string) *Worker {
	return &Worker{
		services:    services,
		baseURL:     parseUrl,
		visaTypeURL: visaTypeUrl,
	}
}

// Run должен быть вызван только после инициализации всех сервисов
func (w *Worker) Run() error {
	// Chat api test
	if err := w.services.Chat.TestConnection(); err != nil {
		return fmt.Errorf("chat api connection error:%w", err)
	}

	// Selenium parse page
	if err := w.services.Parse(w.baseURL); err != nil {
		return fmt.Errorf("page parse error:%w", err)
	}
	log.Println("Web page parsed")

	// Page test
	if err := w.services.Selenium.TestPage(); err != nil {
		return fmt.Errorf("page load test error:%w", err)
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
	if err := w.services.Selenium.GoTo(w.baseURL + w.visaTypeURL); err != nil {
		return fmt.Errorf("go to visa type verification page error:%w", err)
	}

	isAuthorized, _ := w.services.Selenium.IsAuthorized(w.baseURL + w.visaTypeURL)
	if !isAuthorized {
		// Solving first captcha
		if err := w.services.Selenium.ClickVerifyBtn(); err != nil {
			return fmt.Errorf("click verify first catpcha error:%w", err)
		}

		log.Println("Retry process first captcha starts ...")
		if err := w.RetryProcessCaptcha(processCaptchaMaxTries); err != nil {
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

		// TODO: код до "<<<" надо переписать
		// TODO: реализовать ожидание подгрузки следующий страницы (ожидание момента авторизации)
		//DEBUG:
		time.Sleep(time.Second * 5)

		if err := w.services.Selenium.Wd().Get("https://russia.blsspainglobal.com/Global/Bls/VisaTypeVerification"); err != nil {
			return err
		}

		// DEBUG:
		time.Sleep(time.Second * 5)
		// TODO: <<<
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
	if err := w.RetryProcessCaptcha(processCaptchaMaxTries); err != nil {
		return fmt.Errorf("retry process second captcha error:%w", err)
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
		err := w.services.Email.SendAvailbilityNotification(availbilityNotifiedEmail)
		if err != nil {
			return fmt.Errorf("send availability notification error:%w", err)
		}
	} else {
		log.Println("!!!Appointment NOT available!!!")
	}

	// DEBUG:
	time.Sleep(time.Second * 15)

	log.Println("Work done")

	return nil
}

func (w *Worker) LoadCookies() error {
	cookiePath := tmpFolder + cookieFile
	cookiesJson, err := os.ReadFile(cookiePath)
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
	cooks, _ := w.services.Selenium.GetCookies()

	// DEBUG:
	fmt.Println("cookies: ", cooks)

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

	if err := util.WriteFile(cookiePath, cookiesJson); err != nil {
		log.Println("Cannot save cookies:%w", err)
		return
	}

	log.Println("Cookies saved")
}
