package worker

import (
	"fmt"
	"github.com/tebeka/selenium"
	"log"
	"time"
	"visasolution/app/service"
	"visasolution/app/util"
)

const msg = `you see an image with the task: ‘Select all squares with the number …’ Recognize the text in each square and send ONLY the cell numbers that contain this number, separated by commas without spaces. Numbering is left to right.”`

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

	// Work
	log.Println("start solving captcha...")
	if err := w.services.Selenium.ClickButton(selenium.ByCSSSelector, "#btnVerify"); err != nil {
		return fmt.Errorf("click verify error:%w", err)
	}

	if err := w.services.PullCaptchaImage(); err != nil {
		return fmt.Errorf("cannot pull captcha image:%w", err)
	}
	log.Println("captcha was saved")

	link, err := w.services.UploadImage("tmp/captcha.png")
	if err != nil {
		return fmt.Errorf("failed to upload captcha:%w", err)
	}
	log.Println("captcha was uploaded, link: ", link)

	resp, err := w.services.Chat.Request4VPreviewWithImage(msg, link)
	if err != nil {
		return fmt.Errorf("request to chat api with image url error:%w", err)
	}
	cardNums, err := util.StrToIntSlice(w.services.Chat.GetRespMsg(resp), ",")
	log.Println("cards to select: ", cardNums)

	err = w.services.Selenium.ProcessCaptcha(cardNums)
	if err != nil {
		return fmt.Errorf("process captcha error:%w", err)
	}
	log.Println("captcha was sucsessfully processed")

	// Authorization
	if err := w.services.Selenium.Authorize(); err != nil {
		return fmt.Errorf("authorization error:%w", err)
	}

	log.Println("work done")

	time.Sleep(10 * time.Second)

	return nil
}
