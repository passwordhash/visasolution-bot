package service

import (
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"log"
	"visasolution/app/util"
)

type SeleniumService struct {
	wd selenium.WebDriver

	maxTries int
	parseUrl string
}

func NewSeleniumService(maxTries int, parseUrl string) (*SeleniumService, error) {
	wd, err := seleniumConnect("", maxTries)
	if err != nil {
		return &SeleniumService{}, err
	}

	service := &SeleniumService{
		wd:       wd,
		maxTries: maxTries,
		parseUrl: parseUrl,
	}

	return service, nil
}

func (s SeleniumService) Wd() selenium.WebDriver {
	return s.wd
}

func (s SeleniumService) ProcessCaptcha(wd selenium.WebDriver) error {
	var err error

	//elem, err := wd.FindElement(selenium.ByCSSSelector, `#captcha-main-div > div`)
	elem, err := wd.FindElement(selenium.ByXPATH, `//*[@id="popup_1"]/iframe`)
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

func seleniumConnect(url string, maxTries int) (selenium.WebDriver, error) {
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
	for i < maxTries {
		wd, err = selenium.NewRemote(caps, urlPrefix)
		if err != nil {
			log.Println(err)
			i++
			continue
		}
		break
	}

	return wd, err
}
