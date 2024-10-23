package service

import (
	"errors"
	"fmt"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"log"
	"strings"
	"time"
	"visasolution/app/util"
)

const waitDuration = time.Second * 15

const invalidSelectionMsg = "Invalid selection"

var InvalidSelectionError = errors.New("captcha invalid selection")

const (
	captchaIFrameXPath         = `//*[@id="popup_1"]/iframe`
	captchaDragableCSSSelector = `body > div.k-widget.k-window`
	captchaCardsContainerXPath = `//*[@id="captcha-main-div"]/div/div[2]`
	captchaCardImgXPath        = `//*[@id="captcha-main-div"]/div/div[2]/div/img`
	submitCaptchaXPath         = `//*[@id="captchaForm"]/div[2]/div[3]`
	submitCaptchaIdCSSSelector = `#btnSubmit`

	verifyBtnIdCSSSelector = `#btnVerify`

	formInputsXPath            = `/html/body/main/main/div/div/div[2]/div[2]/form/div/input`
	formSubmitId               = `btnSubmit`
	bookNewAppointmentSubmitId = `btnSubmit`
	bookNewFormXPath           = `//*[@id="div-main"]/div/div/div[2]/form`
	bookNewFormControlsXPath   = `//*[@id="div-main"]/div/div/div[2]/form/div`

	bookNewBtnXPath = `//*[@id="tns1-item1"]/div/div/div/div/a`
)

type SeleniumService struct {
	wd selenium.WebDriver

	maxTries int
	parseUrl string

	blsEmail    string
	blsPassword string
}

func NewSeleniumService(maxTries int, blsEmail string, blsPassword string) *SeleniumService {
	return &SeleniumService{maxTries: maxTries, blsEmail: blsEmail, blsPassword: blsPassword}
}

// TODO: объединить функции Connect и ConnectWithProxy
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

// ConnectWithProxy подключается к selenium с прокси аутентификацией.
// url - адрес нашего драйвера
// chromeExtensionPath - путь к расширению для авторизации через прокси
func (s *SeleniumService) ConnectWithProxy(url string, chromeExtensionPath string) error {
	var wd selenium.WebDriver
	var err error

	caps := selenium.Capabilities{
		"browserName": "chrome",
	}

	chrCaps := chrome.Capabilities{
		W3C: true,
	}
	if err := chrCaps.AddExtension(chromeExtensionPath); err != nil {
		return err
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

func (s *SeleniumService) Parse(url string) error {
	return s.wd.Get(url)
}

// TEMP: for easy testing
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

func (s *SeleniumService) DeleteCookie(name string) error {
	return s.wd.DeleteCookie(name)
}

func (s *SeleniumService) MaximizeWindow() error {
	return s.wd.MaximizeWindow("")
}

func (s *SeleniumService) Quit() {
	s.wd.Quit()
}

func (s *SeleniumService) PullCaptchaImage() ([]byte, error) {
	// TODO: REFACTOR перенести часть в worker
	// переключаемся на iframe капчи, находим контейнер, возращаемся обратно,
	// чтобы на скрине было видно содержимое капчи
	var err error

	iframe, err := s.waitAndSwitchIFrame(selenium.ByXPATH, captchaIFrameXPath)
	if err != nil {
		return []byte{}, fmt.Errorf("switch iframe error:%w", err)
	}

	_, err = s.wd.FindElement(selenium.ByCSSSelector, `#captcha-main-div > div`)
	if err != nil {
		return []byte{}, err
	}

	err = s.switchToDefault()
	if err != nil {
		return []byte{}, fmt.Errorf("switch to default frame error:%w", err)
	}

	img, err := iframe.Screenshot(false)
	if err != nil {
		return []byte{}, err
	}

	//return util.WriteFile(util.GetAbsolutePath("tmp/captcha.png"), img)
	return img, nil
}

func (s *SeleniumService) SolveCaptcha(numbers []int) error {
	dragable, err := s.wd.FindElement(selenium.ByCSSSelector, captchaDragableCSSSelector)
	if err != nil {
		return fmt.Errorf("find element 'dragable' error:%w", err)
	}

	// Перемещаем капчу в левый верхний угол
	err = s.changeElementProperties(dragable, map[string]string{
		"left": "0",
		"top":  "0",
	})
	if err != nil {
		return fmt.Errorf("change element property error:%w", err)
	}

	// Переключаемся на фрейм капчи
	_, err = s.waitAndSwitchIFrame(selenium.ByXPATH, captchaIFrameXPath)
	if err != nil {
		return fmt.Errorf("switch iframe error:%w", err)
	}
	defer s.switchToDefault()

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
		// TEMP:
		time.Sleep(time.Millisecond * 500)
		x, y := getCardCoordinates(n, cardW, cardH)
		err = s.clickByCoords(x+horPadding, y+vertPadding)
		if err != nil {
			return fmt.Errorf("click by coords for card number №%d error:%w", n, err)
		}
	}

	time.Sleep(time.Second * 2)

	if err := s.waitAndClickButton(selenium.ByXPATH, submitCaptchaXPath); err != nil {
		return fmt.Errorf("click submit captcha error:%w", err)
	}
	log.Println("submit captcha")

	time.Sleep(time.Second * 2)

	// Обратываем случай неверного решения капчи
	text, err := s.wd.AlertText()
	defer s.wd.AcceptAlert()

	// DEBUG:
	//fmt.Println("alert text and error: ", text, " ", err)
	// TODO: сделать другую проверка на неправилное решение капчи
	if strings.Contains(text, invalidSelectionMsg) || err == nil {
		return InvalidSelectionError
	}

	return nil
}

func (s *SeleniumService) Authorize() error {
	formContols, err := s.wd.FindElements(selenium.ByXPATH, formInputsXPath)
	if err != nil {
		return err
	}

	// Получение только не фейковых input'ов {email, password}
	var controls []selenium.WebElement
	for _, el := range formContols {
		attr, _ := el.GetAttribute("required")
		if attr == "true" {
			controls = append(controls, el)
		}
	}

	err = controls[0].SendKeys(s.blsEmail)
	if err != nil {
		return err
	}
	err = controls[1].SendKeys(s.blsPassword)
	if err != nil {
		return err
	}

	// OLD:
	//time.Sleep(time.Millisecond * 500)
	//return submit.Click()

	err = s.waitAndClickButton(selenium.ByID, formSubmitId)
	if err != nil {
		return fmt.Errorf("submit click error:%w", err)
	}

	return nil
}

// BookNew кликает по кнопке "Book new" на странице. Синхронный метод.
func (s *SeleniumService) BookNew() error {
	err := s.waitAndClickButton(selenium.ByXPATH, bookNewBtnXPath)
	if err != nil {
		return fmt.Errorf("click book new btn error:%w", err)
	}

	return nil
}

func (s *SeleniumService) BookNewAppointment() error {
	// DEBUG:
	time.Sleep(time.Second * 2)
	if err := s.waitAndClickButton(selenium.ByID, bookNewAppointmentSubmitId); err != nil {
		return fmt.Errorf("submit to book new appointment error:%w", err)
	}

	// Удостоверяемся, что форма для бронирования новой записи загрузилась
	_, err := s.waitAndFind(selenium.ByXPATH, bookNewFormXPath)
	if err != nil {
		return fmt.Errorf("find 'book new' form error:%w", err)
	}

	// Находим все нефейковые элементы формы
	formContols, err := s.wd.FindElements(selenium.ByXPATH, bookNewFormControlsXPath)
	if err != nil {
		return fmt.Errorf("find 'book new' form controls error:%w", err)
	}

	log.Printf("len of form: %d\n", len(formContols))

	formContolsDisplayed := make([]selenium.WebElement, 0, len(formContols))
	// Начинаем с 2 элемента, т.к. первые 2 элемента - это заголовок и текст
	for _, el := range formContols[2:len(formContols)] {
		displayed, err := el.IsDisplayed()
		if err != nil {
			//DEBUG:
			fmt.Println("check displayed error: ", err)
			return fmt.Errorf("check displayed error:%w", err)
		}
		if displayed {
			formContolsDisplayed = append(formContolsDisplayed, el)
		}
	}

	log.Printf("len of form displ: %d\n", len(formContols))

	// Заполняем все необходимые поля
	for i, el := range formContolsDisplayed {
		log.Printf("ELEMENT %d:\n", i+1)
		fmt.Println("el: ", el)

		input, err := el.FindElement(selenium.ByTagName, "input")
		if err != nil {
			return fmt.Errorf("find input error:%w", err)
		}

		id, err := input.GetAttribute("id")
		if err != nil {
			return fmt.Errorf("get input id error:%w", err)
		}

		log.Println("id: ", id)
	}

	return nil
}

// ClickVerifyBtn кликает по кнопке с ожиданием появления элемента. Синхронный метод.
func (s *SeleniumService) ClickVerifyBtn() error {
	return s.waitAndClickButton(selenium.ByCSSSelector, verifyBtnIdCSSSelector)
}

// clickButton кликает по элементу по заданным параметрам
func (s *SeleniumService) clickButton(byWhat, value string) error {
	elem, err := s.wd.FindElement(byWhat, value)
	if err != nil {
		return err
	}

	return elem.Click()
}

// TODO: refactor all wait funcs
// waitAndFind ожидает появления элемента и возвращает его
func (s *SeleniumService) waitAndFind(byWhat, value string) (selenium.WebElement, error) {
	var element selenium.WebElement
	var err error
	maxTries := 10                  // HARD CODED
	delay := time.Millisecond * 500 // HARD CODED

	for i := 0; i < maxTries; i++ {
		element, err = s.wd.FindElement(byWhat, value)
		if err == nil {
			return element, nil
		}
		time.Sleep(delay)
	}

	return nil, fmt.Errorf("element not found after multiple attempts: %w", err)
}

// waitAndClickButton ожидает появления элемента и кликает по нему
func (s *SeleniumService) waitAndClickButton(byWhat, value string) error {
	var err error
	maxTries := 10                  // HARD CODED
	delay := time.Millisecond * 500 // HARD CODED
	for i := 0; i < maxTries; i++ {
		err = s.clickButton(byWhat, value)
		if err == nil {
			return nil
		}
		// DEBUG:
		fmt.Println("wait and click error:", err)
		time.Sleep(delay)
	}
	return fmt.Errorf("max tries exceeded:%w", err)
}

// waitAndSwitchIFrame ожидает появления элемента и переключается на него
func (s *SeleniumService) waitAndSwitchIFrame(byWhat, value string) (selenium.WebElement, error) {
	var iframe selenium.WebElement
	var err error
	maxTries := 10 // HARD CODED
	//delay := time.Millisecond * 500 // HARD CODED
	delay := time.Second * 2

	for i := 0; i < maxTries; i++ {
		iframe, err = s.wd.FindElement(byWhat, value)
		if err == nil {
			err = s.wd.SwitchFrame(iframe)
			if err == nil {
				return iframe, nil
			}
		}
		time.Sleep(delay)
	}

	return nil, fmt.Errorf("element not found after multiple attempts: %w", err)
}

// switchToDefault переключается на дефолтный фрейм
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
