package main

import (
	"fmt"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"log"
	"os"
	"time"
	cfg "visasolution/config"
)

//docker run --rm -p=4444:4444 selenium/standalone-chrome

const maxTries = 10
const parseURL = "https://russia.blsspainglobal.com/Global/account/login"

func main() {
	// load confgi
	config, err := cfg.LoadConfig()
	if err != nil {
		log.Fatalln(err)
	}

	// selenium connect
	wd, err := connect()
	if err != nil {
		log.Print("web driver connection error: ", err)
		return
	}
	defer wd.Quit()
	log.Println("web driver connected")

	if err = wd.MaximizeWindow(""); err != nil {
		log.Println("maximizing window error: ", err)
	}

	getURL := parseURL
	err = wd.Get(getURL)
	if err != nil {
		log.Println("get error: ", err)
		return
	}

	// test
	elem, err := wd.FindElement(selenium.ByXPATH, "//*[@id=\"navbarCollapse\"]/div[1]/div/div/div/div")
	if err != nil {
		log.Println("DOM test failed: ", err)
		return
	}

	text, err := elem.Text()
	if err != nil {
		log.Println("cannot get text of element: ", elem)
		return
	}

	log.Println("finded: ", text)

	// form
	emailInput, err := wd.FindElement(selenium.ByXPATH, "//*[@id=\"UserId6\"]")
	if err != nil {
		log.Println("cannot get email input element: ", err)
		return
	}

	passwordInput, err := wd.FindElement(selenium.ByXPATH, "//*[@id=\"Password2\"]")
	if err != nil {
		log.Println("cannot get password input element: ", err)
		return
	}

	verifyBtn, err := wd.FindElement(selenium.ByXPATH, "//*[@id=\"btnVerify\"]")
	if err != nil {
		log.Println("cannot get verify btn: ", err)
		return
	}

	time.Sleep(time.Second * 10)
	log.Println(verifyBtn.Click())
	time.Sleep(time.Second * 10)

	err = processCaptcha(wd)
	if err != nil {
		log.Println("error while proccessing captcha: ", err)
		return
	}

	time.Sleep(time.Second * 10)

	// FIXME: not iteractable
	err = emailInput.SendKeys(config.BlsEmail)
	if err != nil {
		log.Println("cannot send key in email input: ", err)
		return
	}

	// FIXME: not iteractable
	err = passwordInput.SendKeys(config.BlsPassword)
	if err != nil {
		log.Println("cannot send key in password input: ", err)
		return
	}

	fmt.Println("finish")
}

func processCaptcha(wd selenium.WebDriver) error {
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

	file, err := os.Create("hello.png")
	if err != nil {
		return err
	}
	_, err = file.Write(img)

	return err
}

func connect() (selenium.WebDriver, error) {
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
	urlPrefix := selenium.DefaultURLPrefix
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
