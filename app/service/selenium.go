package service

import (
	"fmt"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"log"
	"time"
	"visasolution/app/util"
)

const waitDuration = time.Second * 15

type SeleniumService struct {
	wd selenium.WebDriver

	maxTries int
	parseUrl string
}

func NewSeleniumService(maxTries int) *SeleniumService {
	return &SeleniumService{
		maxTries: maxTries,
	}
}

func (s *SeleniumService) Parse(url string) error {
	return s.wd.Get(url)
}

func (s *SeleniumService) Wd() selenium.WebDriver {
	return s.wd
}

func (s *SeleniumService) MaximizeWindow() error {
	return s.wd.MaximizeWindow("")
}

func (s *SeleniumService) ProcessCaptcha() error {
	var err error

	err = s.clickButton(selenium.ByCSSSelector, "#btnVerify")
	if err != nil {
		return fmt.Errorf("click verify error:%w", err)
	}

	// капча находится в iframe, нужно найти и переключиться на него
	var iframe selenium.WebElement
	err = s.wd.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		iframe, err = s.wd.FindElement(selenium.ByXPATH, `//*[@id="popup_1"]/iframe`)
		if err != nil {
			return false, err
		}

		return true, nil
	}, waitDuration)
	if err != nil {
		return err
	}

	err = s.wd.SwitchFrame(iframe)
	if err != nil {
		return err
	}
	defer s.wd.SwitchFrame(nil) // переключение обратно в корневой документ

	captcha, err := s.wd.FindElement(selenium.ByCSSSelector, `#captcha-main-div > div`)
	if err != nil {
		return err
	}

	img, err := captcha.Screenshot(false)
	if err != nil {
		return err
	}

	return util.WriteFile(util.GetAbsolutePath("tmp/captcha.png"), img)
}

func (s *SeleniumService) clickButton(byWhat, value string) error {
	elem, err := s.wd.FindElement(byWhat, value)
	if err != nil {
		return err
	}

	return elem.Click()
}

//func (s *SeleniumService) ProcessCaptcha() error {
//	var err error
//
//	elem, err := wd.FindElement(selenium.ByCSSSelector, `#captcha-main-div > div`)
//	elem, err := s.wd.FindElement(selenium.ByXPATH, `//*[@id="popup_1"]/iframe`)
//	if err != nil {
//		return err
//	}
//	img, err := elem.Screenshot(false)
//	if err != nil {
//		return err
//	}
//
//	err = util.WriteFile(util.GetAbsolutePath("tmp/captcha.png"), img)
//
//	return err
//}

func (s *SeleniumService) Connect(url string) error {
	var wd selenium.WebDriver
	var err error

	caps := selenium.Capabilities{
		"browserName": "chrome",
	}

	chrCaps := chrome.Capabilities{
		W3C: true,
	}
	caps.AddChrome(chrCaps)

	// адрес нашего драйвера
	urlPrefix := url
	if url == "" {
		urlPrefix = selenium.DefaultURLPrefix
	}
	i := 0
	for i < s.maxTries {
		wd, err = selenium.NewRemote(caps, urlPrefix)
		if err != nil {
			log.Println(err)
			i++
			time.Sleep(time.Second * 1)
			continue
		}
		break
	}

	s.wd = wd

	return err
}

func (s *SeleniumService) Quit() {
	s.wd.Quit()
}
