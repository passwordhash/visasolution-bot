package worker

import (
	"fmt"
	"log"
	"time"
	"visasolution/app/service"
)

const processCaptchaMaxTries = 3

const tmpFolder = "tmp/"

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
		return fmt.Errorf("retry process captcha error:%w", err)
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
	time.Sleep(time.Second * 15)

	//if err := w.services.Wd().Refresh(); err != nil {
	//	return fmt.Errorf("cannot refresh: %w", w)
	//}
	if err := w.services.Selenium.Wd().Get("https://russia.blsspainglobal.com/Global/Bls/VisaTypeVerification"); err != nil {
		return err
	}

	// DEBUG:
	time.Sleep(time.Second * 15)

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

	log.Println("Retry process second captcha starts ...")
	if err := w.RetryProcessCaptcha(processCaptchaMaxTries); err != nil {
		return fmt.Errorf("retry process captcha error:%w", err)
	}
	log.Println("Retry process second captcha successfully ended")

	// DEBUG:
	time.Sleep(time.Second * 10)

	log.Println("Work done")

	return nil
}
