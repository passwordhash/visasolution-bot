package service

import (
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"log"
	"time"
	"visasolution/app/util"
)

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

func (s *SeleniumService) Wd() selenium.WebDriver {
	return s.wd
}

func (s *SeleniumService) MaximizeWindow() error {
	return s.wd.MaximizeWindow("")
}

func (s *SeleniumService) ProcessCaptcha() error {
	var err error

	//elem, err := wd.FindElement(selenium.ByCSSSelector, `#captcha-main-div > div`)
	elem, err := s.wd.FindElement(selenium.ByXPATH, `//*[@id="popup_1"]/iframe`)
	if err != nil {
		return err
	}
	img, err := elem.Screenshot(false)
	if err != nil {
		return err
	}

	err = util.WriteFile(util.GetAbsolutePath("tmp/captcha.png"), img)

	return err
}

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
