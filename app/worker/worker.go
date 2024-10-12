package worker

import (
	"fmt"
	"log"
	"visasolution/app/service"
)

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

// Run function should be called when selenium is already connected and chat api client is inited
func (w *Worker) Run() error {
	// Chat api test
	if err := w.services.Chat.TestConnection(); err != nil {
		return fmt.Errorf("chat api connection error:%w", err)
	}

	// Selenium parse page
	if err := w.services.Parse(w.parseUrl); err != nil {
		return fmt.Errorf("page parse error:%w", err)
	}
	log.Println("web page parsed")

	// Page test
	if err := w.services.Selenium.TestPage(); err != nil {
		return fmt.Errorf("page load test error:%w", err)
	}
	log.Println("page successfully loaded")

	if err := w.services.Selenium.MaximizeWindow(); err != nil {
		return fmt.Errorf("cannot maximize window:%w", err)
	}

	//time.Sleep(10 * time.Second)

	if err := w.services.ProcessCaptcha(); err != nil {
		return fmt.Errorf("cannot process captcha:%w", err)
	}
	log.Println("captcha was saved")

	link, err := w.services.UploadImage("tmp/captcha.png")
	if err != nil {
		return fmt.Errorf("failed to upload captcha:%w", err)
	}
	log.Println("captcha was uploaded, link: ", link)

	return nil
}
