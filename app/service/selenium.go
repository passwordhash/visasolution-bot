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

const (
	captchaIFrameXPath         = `//*[@id="popup_1"]/iframe`
	captchaDragableCSSSelector = `body > div.k-widget.k-window`
	captchaCardsContainerXPath = `//*[@id="captcha-main-div"]/div/div[2]`
	captchaCardImgXPath        = `//*[@id="captcha-main-div"]/div/div[2]/div/img`
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

func (s *SeleniumService) Parse(url string) error {
	return s.wd.Get(url)
}

func (s *SeleniumService) Wd() selenium.WebDriver {
	return s.wd
}

func (s *SeleniumService) TestPage() error {
	var err error

	// header
	_, err = s.wd.FindElement(selenium.ByXPATH, `/html/body/header/nav[1]`)
	// body
	_, err = s.wd.FindElement(selenium.ByXPATH, `//*[@id="div-main"]`)
	// footer
	_, err = s.wd.FindElement(selenium.ByXPATH, `/html/body/footer/div/div[1]/div[1]/h4`)

	return err
}

func (s *SeleniumService) MaximizeWindow() error {
	return s.wd.MaximizeWindow("")
}

func (s *SeleniumService) PullCaptchaImage() error {
	// переключаемся на iframe капчи, находим контейнер, возращаемся обратно,
	// чтобы на скрине было видно содержимое капчи
	var err error

	iframe, err := s.switchIFrame(selenium.ByXPATH, captchaIFrameXPath)
	if err != nil {
		return fmt.Errorf("switch iframe error:%w", err)
	}

	_, err = s.wd.FindElement(selenium.ByCSSSelector, `#captcha-main-div > div`)
	if err != nil {
		return err
	}

	err = s.switchToDefault()
	if err != nil {
		return fmt.Errorf("switch to default frame error:%w", err)
	}

	img, err := iframe.Screenshot(false)
	if err != nil {
		return err
	}

	return util.WriteFile(util.GetAbsolutePath("tmp/captcha.png"), img)
}

func (s *SeleniumService) ProcessCaptcha(numbers []int) error {
	dragable, err := s.wd.FindElement(selenium.ByCSSSelector, captchaDragableCSSSelector)
	if err != nil {
		return err
	}

	// Перемещаем качу в левых верхний угол
	err = s.changeElementProperties(dragable, map[string]string{
		"left": "0",
		"top":  "0",
	})
	if err != nil {
		return fmt.Errorf("change element property error:%w", err)
	}

	// Получаем размеры квадрата
	_, err = s.switchIFrame(selenium.ByXPATH, captchaIFrameXPath)
	if err != nil {
		return fmt.Errorf("switch iframe error:%w", err)
	}
	defer s.switchToDefault()

	cardImg, err := s.wd.FindElement(selenium.ByXPATH, captchaCardImgXPath)
	if err != nil {
		return err
	}

	width, height, err := s.getElementSizes(cardImg)
	if err != nil {
		return fmt.Errorf("getting elemnt sizes error:%w", err)
	}

	log.Println("card image sizes: ", width, " ", height)

	time.Sleep(10 * time.Second)

	return nil
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

func (s *SeleniumService) ClickButton(byWhat, value string) error {
	elem, err := s.wd.FindElement(byWhat, value)
	if err != nil {
		return err
	}

	return elem.Click()
}

func (s *SeleniumService) switchIFrame(byWhat, value string) (selenium.WebElement, error) {
	var iframe selenium.WebElement
	var err error
	err = s.wd.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		iframe, err = s.wd.FindElement(byWhat, value)
		if err != nil {
			return false, err
		}
		return true, nil
	}, waitDuration)
	if err != nil {
		return iframe, err
	}

	err = s.wd.SwitchFrame(iframe)
	if err != nil {
		return iframe, err
	}

	return iframe, err
}

func (s *SeleniumService) switchToDefault() error {
	return s.wd.SwitchFrame(nil)
}

func (s *SeleniumService) changeElementProperties(elem selenium.WebElement, props map[string]string) error {
	var script string
	for k, v := range props {
		script += fmt.Sprintf("arguments[0].style.%s = '%s';\n", k, v)
	}
	_, err := s.wd.ExecuteScript(script, []interface{}{elem})
	return err
}

func (s *SeleniumService) getElementSizes(elem selenium.WebElement) (int, int, error) {
	var width, heigth int
	var err error

	widthProp, err := elem.CSSProperty("width")
	if err != nil {
		return 0, 0, err
	}

	width, err = util.PxToInt(widthProp)
	if err != nil {
		return 0, 0, fmt.Errorf("convert px to int error:%w", err)
	}

	heightProp, err := elem.CSSProperty("height")
	if err != nil {
		return 0, 0, err
	}

	heigth, err = util.PxToInt(heightProp)
	if err != nil {
		return 0, 0, fmt.Errorf("convert px to int error:%w", err)
	}

	return width, heigth, nil
}
