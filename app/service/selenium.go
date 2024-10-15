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

	submitCaptchaXPath = `//*[@id="captchaForm"]/div[2]/div[3]`
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

	// Перемещаем капчу в левых верхний угол
	err = s.changeElementProperties(dragable, map[string]string{
		"left": "0",
		"top":  "0",
	})
	if err != nil {
		return fmt.Errorf("change element property error:%w", err)
	}

	// Переключаемся на фрейм капчи
	_, err = s.switchIFrame(selenium.ByXPATH, captchaIFrameXPath)
	if err != nil {
		return fmt.Errorf("switch iframe error:%w", err)
	}
	defer s.switchToDefault()

	submitBtn, err := s.wd.FindElement(selenium.ByXPATH, submitCaptchaXPath)
	if err != nil {
		return err
	}

	// Получаем размеры карточек
	cardImg, err := s.wd.FindElement(selenium.ByXPATH, captchaCardImgXPath)
	if err != nil {
		return err
	}

	cardW, cardH, err := s.getElementSizes(cardImg)
	if err != nil {
		return fmt.Errorf("getting card image sizes error:%w", err)
	}

	// Вычисления в примерных значения
	horPadding := cardW / 2
	vertPadding := int(float64(cardH) * 1.15)

	// Проходимся по номерам карточек и кликаем вычисленным координатам для каждого номера
	for _, n := range numbers {
		x, y := getCardCoordinates(n, cardW, cardH)
		err = s.clickByCoords(x+horPadding, y+vertPadding)
		if err != nil {
			return fmt.Errorf("click by coords for card number №%d error:%w", n, err)
		}
	}

	time.Sleep(10 * time.Second)

	if err := submitBtn.Click(); err != nil {
		return err
	}

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

func (s *SeleniumService) clickByCoords(x, y int) error {
	script := `
    var event = new MouseEvent('click', {
        'view': window,
        'bubbles': true,
        'cancelable': true,
        'clientX': arguments[0],
        'clientY': arguments[1]
    });
    document.elementFromPoint(arguments[0], arguments[1]).dispatchEvent(event);
`
	_, err := s.wd.ExecuteScript(script, []interface{}{x, y})
	return err
}

func getCardCoordinates(cardNum, cardWidth, cardHeight int) (int, int) {
	row := (cardNum - 1) / 3
	col := (cardNum - 1) % 3

	x := col*cardWidth + cardWidth/2
	y := row*cardHeight + cardHeight/2

	return x, y
}
