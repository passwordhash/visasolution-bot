package service

import (
	"errors"
	"fmt"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"log"
	"strings"
	"time"
	util2 "visasolution/pkg/util"
)

const authCookieKey = ".AspNetCore.Cookies"

// Константы для нахождения элементов на странице
const (
	captchaIFrameXPath         = `//*[@id="popup_1"]/iframe`
	captchaDragableCSSSelector = `body > div.k-widget.k-window`
	captchaCardImgXPath        = `//*[@id="captcha-main-div"]/div/div[2]/div/img`
	submitCaptchaXPath         = `//*[@id="captchaForm"]/div[2]/div[3]`

	verifyBtnIdCSSSelector = `#btnVerify`

	formInputsXPath            = `/html/body/main/main/div/div/div[2]/div[2]/form/div/input`
	formSubmitId               = `btnSubmit`
	bookNewAppointmentId       = `btnSubmit`
	bookNewAppointmentSubmitId = `btnSubmit`
	bookNewFormXPath           = `//*[@id="div-main"]/div/div/div[2]/form`
	bookNewFormControlsXPath   = `//*[@id="div-main"]/div/div/div[2]/form/div`

	commonModalId       = `commonModal`
	commonModalHeaderId = `commonModalHeader`

	bookNewBtnXPath = `//*[@id="tns1-item1"]/div/div/div/div/a`
)

const (
	availabilityCheckMsg = "No Appointments Available"
	invalidSelectionMsg  = "Invalid selection"
)

// SeleniumLegacyCode тип для легаси кодов ошибок Selenium WebDriver
type SeleniumLegacyCode int

const (
	invalidSessionId = iota
)

var InvalidSelectionError = errors.New("captcha invalid selection")

var InvalidSessionError = errors.New("invalid session id")

// Количество табов до первого input'а в форме Book New Appointment
const tabsCountToFirstInput = 14

// Нужное количество нажатий на стрелку вниз для выбора каждого элемента формы по id. -1 означает, что элемент не нужно выбирать
var inputKeyDownCounts = map[string]int{
	"self":                  -1,
	"ApplicantsNo":          -1,
	"JurisdictionId":        1,
	"loc":                   4,
	"VisaType":              1,
	"VisaSubType":           9,
	"AppointmentCategoryId": 1,
}

type SeleniumService struct {
	wd selenium.WebDriver

	maxTries    int
	seleniumURL string
	blsEmail    string
	blsPassword string
}

func NewSeleniumService(maxTries int, blsEmail string, seleniumURL string, blsPassword string) *SeleniumService {
	return &SeleniumService{
		maxTries:    maxTries,
		blsEmail:    blsEmail,
		seleniumURL: seleniumURL,
		blsPassword: blsPassword,
	}
}

// ConnectWithProxy подключается к selenium с прокси аутентификацией.
// url - адрес нашего драйвера
// chromeExtensionPath - путь к расширению для авторизации через прокси
func (s *SeleniumService) ConnectWithProxy(chromeExtensionPath string) error {
	var wd selenium.WebDriver

	caps := selenium.Capabilities{
		"browserName": "chrome",
	}

	chrCaps := chrome.Capabilities{
		W3C: true,
	}
	err := chrCaps.AddExtension(chromeExtensionPath)
	if err != nil {
		return err
	}
	caps.AddChrome(chrCaps)

	urlPrefix := s.seleniumURL
	if urlPrefix == "" {
		urlPrefix = selenium.DefaultURLPrefix
	}
	for i := 0; i < s.maxTries; i++ {
		wd, err = selenium.NewRemote(caps, urlPrefix)
		if err == nil {
			break
		}
		log.Println(err)
		time.Sleep(3 * time.Second)
	}

	s.wd = wd

	return err
}

func (s *SeleniumService) TestPage() error {
	if _, err := s.wd.FindElement(selenium.ByXPATH, `/html/body/header/`); err != nil {
		return err
	}
	if _, err := s.wd.FindElement(selenium.ByXPATH, `//*[@id="div-main"]`); err != nil {
		return err
	}
	if _, err := s.wd.FindElement(selenium.ByXPATH, `/html/body/footer/`); err != nil {
		return err
	}
	return nil
}

// GoTo переходит на страницу по url
func (s *SeleniumService) GoTo(url string) error {
	err := s.wd.Get(url)
	var seleniumErr *selenium.Error
	if errors.As(err, &seleniumErr) {
		if seleniumErr.LegacyCode == invalidSessionId || seleniumErr.Err == InvalidSessionError.Error() {
			return InvalidSessionError
		}
		return err
	}

	return nil
}

// IsAuthorized проверяет авторизован ли пользователь, но только для одной конкретной страницы - проверка по URL - VisaTypeVerification
// TODO: сделать общий метод для проверки авторизации
func (s *SeleniumService) IsAuthorized(neededURL string) (bool, error) {
	curURL, err := s.wd.CurrentURL()
	if err != nil {
		return false, err
	}

	return curURL == neededURL, nil
}

// AuthCookie возвращает куки авторизации
func (s *SeleniumService) AuthCookie() (selenium.Cookie, error) {
	return s.wd.GetCookie(authCookieKey)
}

func (s *SeleniumService) Cookies() ([]selenium.Cookie, error) {
	return s.wd.GetCookies()
}

func (s *SeleniumService) SetCookies(cookies []selenium.Cookie) error {
	for _, c := range cookies {
		if err := s.wd.AddCookie(&c); err != nil {
			return err
		}
	}
	return nil
}

func (s *SeleniumService) DeleteCookie(name string) error {
	return s.wd.DeleteCookie(name)
}

func (s *SeleniumService) DeleteAllCookies() error {
	return s.wd.DeleteAllCookies()
}

func (s *SeleniumService) MaximizeWindow() error {
	return s.wd.MaximizeWindow("")
}

func (s *SeleniumService) Refresh() error {
	return s.wd.Refresh()
}

func (s *SeleniumService) Quit() error {
	return s.wd.Quit()
}

// PullPageScreenshot возвращает скриншот страницы в виде среза байт
func (s *SeleniumService) PullPageScreenshot() ([]byte, error) {
	return s.wd.Screenshot()
}

// PullCaptchaImage возвращает изображение капчи в виде среза байт
func (s *SeleniumService) PullCaptchaImage() ([]byte, error) {
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

	return img, nil
}

// SolveCaptcha проходит уже решенную капчу. На вход принимает срез номеров карточек с 1 по 9
// TODO: сделать возращение bool, свидетельствующее о том, что капча решена/не решена
func (s *SeleniumService) SolveCaptcha(numbers []int) error {
	dragable, err := s.wd.FindElement(selenium.ByCSSSelector, captchaDragableCSSSelector)
	if err != nil {
		return fmt.Errorf("find element 'dragable' error:%w", err)
	}

	err = s.changeElementProperties(dragable, map[string]string{
		"left": "0",
		"top":  "0",
	})
	if err != nil {
		return fmt.Errorf("change element property error:%w", err)
	}

	_, err = s.waitAndSwitchIFrame(selenium.ByXPATH, captchaIFrameXPath)
	if err != nil {
		return fmt.Errorf("switch iframe error:%w", err)
	}
	defer s.switchToDefault()

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
		time.Sleep(time.Millisecond * 200)
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

	text, err := s.wd.AlertText()
	defer s.wd.AcceptAlert()
	// TODO: сделать другую проверка на неправилное решение капчи
	if err == nil || strings.Contains(text, invalidSelectionMsg) {
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

// BookNewAppointment заполняет форму "Book New Appointment" и отправляет ее
func (s *SeleniumService) BookNewAppointment() error {
	if err := s.waitAndClickButton(selenium.ByID, bookNewAppointmentId); err != nil {
		return fmt.Errorf("submit to book new appointment error: %w", err)
	}

	if _, err := s.waitAndFind(selenium.ByXPATH, bookNewFormXPath); err != nil {
		return fmt.Errorf("find 'book new' form error: %w", err)
	}

	// TODO: сделать ожидание появления элементов формы
	time.Sleep(time.Second * 3)

	formControlsDisplayed, err := s.getDisplayedFormControls()
	if err != nil {
		return fmt.Errorf("get displayed form control items error: %w", err)
	}

	if err := s.keyDownFor(tabsCountToFirstInput, selenium.TabKey); err != nil {
		return fmt.Errorf("move to first input error: %w", err)
	}

	for _, el := range formControlsDisplayed {
		time.Sleep(time.Millisecond * 300)

		input, err := el.FindElement(selenium.ByTagName, "input")
		if err != nil {
			continue
		}

		id, err := input.GetAttribute("id")
		if err != nil {
			return fmt.Errorf("get input id error: %w", err)
		}
		// HARD CODED:
		if strings.Contains(id, "ApplicantsNo") {
			continue
		}

		sanitizedId := util2.WithoutDigits(id)

		keyDownCount := inputKeyDownCounts[sanitizedId]
		if err := s.keyDownFor(keyDownCount, selenium.DownArrowKey); err != nil {
			return fmt.Errorf("press arrow down error: %w", err)
		}

		if err := s.wd.KeyDown(selenium.TabKey); err != nil {
			return fmt.Errorf("press tabkey down error: %w", err)
		}
	}

	if err := s.waitAndClickButton(selenium.ByID, bookNewAppointmentSubmitId); err != nil {
		return fmt.Errorf("submit book new appointment form error: %w", err)
	}

	return nil
}

// CheckAvailability проверяет доступность регистрации на получение визы
func (s *SeleniumService) CheckAvailability() (bool, error) {
	commonModal, err := s.waitAndFind(selenium.ByID, commonModalId)
	if err != nil {
		return false, fmt.Errorf("find common modal error: %w", err)
	}

	isDisplayed, err := commonModal.IsDisplayed()
	if err != nil {
		return false, fmt.Errorf("check common modal displayed error: %w", err)
	}

	header, err := commonModal.FindElement(selenium.ByID, commonModalHeaderId)
	if err != nil {
		return false, fmt.Errorf("find common modal header error: %w", err)
	}

	text, err := header.Text()
	if err != nil {
		return false, fmt.Errorf("get common modal header text error: %w", err)
	}

	isAvailable := !(strings.Contains(text, availabilityCheckMsg) || isDisplayed)

	return isAvailable, nil
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
	maxTries := s.maxTries          // HARD CODED
	delay := time.Millisecond * 500 // HARD CODED
	for i := 0; i < maxTries; i++ {
		err = s.clickButton(byWhat, value)
		if err == nil {
			return nil
		}
		time.Sleep(delay)
	}
	return fmt.Errorf("max tries exceeded:%w", err)
}

// waitAndSwitchIFrame ожидает появления IFrame'а и переключается на него
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

// switchToDefault переключается на дефолтный фрейм (основной html документ)
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

	width, err = util2.PxToInt(widthProp)
	if err != nil {
		return 0, 0, fmt.Errorf("convert px to int error:%w", err)
	}

	heightProp, err := elem.CSSProperty("height")
	if err != nil {
		return 0, 0, err
	}

	heigth, err = util2.PxToInt(heightProp)
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

// keyDownFor нажимает key клавишу times раз
func (s *SeleniumService) keyDownFor(times int, key string) error {
	for i := 0; i < times; i++ {
		if err := s.wd.KeyDown(key); err != nil {
			return err
		}
	}
	return nil
}

// getDisplayedFormControls возвращает только отображаемые элементы формы.
// Только для страницы Book New Appointment (форма для проверки доступности записи)
func (s *SeleniumService) getDisplayedFormControls() ([]selenium.WebElement, error) {
	formControls, err := s.wd.FindElements(selenium.ByXPATH, bookNewFormControlsXPath)
	if err != nil {
		return nil, fmt.Errorf("find 'book new' form controls error: %w", err)
	}

	if len(formControls) < 2 {
		return nil, errors.New("no form controls found")
	}

	formControlsDisplayed := make([]selenium.WebElement, 0)
	for _, el := range formControls[2:] {
		displayed, err := el.IsDisplayed()
		if err != nil {
			return nil, fmt.Errorf("check displayed error: %w", err)
		}
		if displayed {
			formControlsDisplayed = append(formControlsDisplayed, el)
		}
	}

	return formControlsDisplayed, nil
}

func getCardCoordinates(cardNum, cardWidth, cardHeight int) (int, int) {
	row := (cardNum - 1) / 3
	col := (cardNum - 1) % 3

	x := col*cardWidth + cardWidth/2
	y := row*cardHeight + cardHeight/2

	return x, y
}
