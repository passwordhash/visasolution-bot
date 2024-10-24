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

const processCaptchaMaxTries = 3

const (
	tmpFolder  = "tmp/"
	cookieFile = "cookies.json"
)

var cookiePath = tmpFolder + cookieFile

type Worker struct {
	services *service.Service
	parseUrl string
}

func NewWorker(services *service.Service, parseUrl string) *Worker {
	return &Worker{
		services: services,
		parseUrl: parseUrl,
	}
}

// Run должен быть вызван только после инициализации всех сервисов
func (w *Worker) Run() error {
	if err := w.services.Wd().Get("https://russia.blsspainglobal.com/Global/Bls/VisaTypeVerification"); err != nil {
		return err
	}

	time.Sleep(time.Second * 10)

	return nil
	// Chat api test
	if err := w.services.Chat.TestConnection(); err != nil {
		return fmt.Errorf("chat api connection error:%w", err)
	}

	// Selenium parse page
	if err := w.services.Parse(w.parseUrl); err != nil {
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

	// Delete auth cookie
	if err := w.services.Selenium.DeleteCookie(".AspNetCore.Cookies"); err != nil {
		return err
	}

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
	time.Sleep(time.Second * 5)

	//if err := w.services.Wd().Refresh(); err != nil {
	//	return fmt.Errorf("cannot refresh: %w", w)
	//}
	if err := w.services.Selenium.Wd().Get("https://russia.blsspainglobal.com/Global/Bls/VisaTypeVerification"); err != nil {
		return err
	}

	// DEBUG:
	time.Sleep(time.Second * 20)

	// Book new
	//if err := w.services.Selenium.BookNew(); err != nil {
	//	return fmt.Errorf("book new error:%w", err)
	//}
	// or:
	//if err := w.services.Selenium.Wd().Get("https://russia.blsspainglobal.com/Global/Bls/VisaTypeVerification"); err != nil {
	//	return err
	//}
	// TODO: <<<

	// Solving second captcha
	if err := w.services.Selenium.ClickVerifyBtn(); err != nil {
		return fmt.Errorf("click verify second captcha error:%w", err)
	}
	log.Println("Second captcha successfully clicked")

	// DEBUG:
	time.Sleep(time.Second * 5)

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
	time.Sleep(time.Second * 15)

	log.Println("Work done")

	return nil
}

// SaveCookies сохраняет куки в файл. Функция вызывается в defer
func (w *Worker) SaveCookies() {
	cookies, err := w.services.Selenium.GetCookies()
	if err != nil {
		log.Println("Cannot get cookies:%w", err)
		return
	}

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

	if err := w.services.Selenium.SetCookies(cookies); err != nil {
		return fmt.Errorf("cannot set cookies:%w", err)
	}

	return nil
}
