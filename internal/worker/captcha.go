package worker

import (
	"errors"
	"fmt"
	"log"
	"visasolution/internal/service"
	util "visasolution/pkg/util"
)

const (
	// msg сообщение, которое отправляется chat api
	msg        = `you see an image with the task: ‘Select all squares with the number …’ Recognize the text in each square and send ONLY the cell numbers that contain this number, separated by commas without spaces. Numbering is left to right starting with 1.Take your time when choosing cards. The wrong decision is costly ”`
	captchaImg = "captcha.png"
)

// RetryProcessCaptcha пытается решить капчу заданное количество раз
func (w *Worker) RetryProcessCaptcha(maxTries int) error {
	for cntTries := 1; cntTries <= maxTries; cntTries++ {
		log.Printf("try No %d to solve the captcha starts ...\n", cntTries)
		err := w.processCaptcha()
		log.Printf("try No %d to solve the captcha ended\n", cntTries)
		if err == nil {
			return nil
		}
		if errors.Is(err, service.InvalidSelectionError) {
			log.Println("invalid selection error, try again")
			continue
		}
		return fmt.Errorf("solve captcha error in try №%d:%w\n", cntTries, err)
	}
	return fmt.Errorf("couldnt solve captcha after %d tries", maxTries)
}

// processCaptcha обрабатывает капчу, занимается ее решением
func (w *Worker) processCaptcha() error {
	err := w.saveCaptchaImage(w.captchaImgPath())
	if err != nil {
		return fmt.Errorf("save captcha image error:%w", err)
	}

	link, err := w.services.UploadImage(w.captchaImgPath())
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

	err = w.services.Selenium.SolveCaptcha(cardNums)
	if err != nil {
		return err
	}

	return nil
}

// saveCaptchaImage сохраняет изображение капчи
func (w *Worker) saveCaptchaImage(relativePath string) error {
	img, err := w.services.Selenium.PullCaptchaImage()
	if err != nil {
		return fmt.Errorf("cannot pull captcha image:%w", err)
	}
	return util.WriteFile(relativePath, img)
}

func (w *Worker) captchaImgPath() string {
	return w.d.TmpFolder + captchaImg
}
