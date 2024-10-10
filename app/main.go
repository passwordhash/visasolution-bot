package main

import (
	"fmt"
	"github.com/tebeka/selenium"
	"log"
	"time"
	cfg "visasolution/app/config"
	"visasolution/app/service"
)

//docker run --rm -p=4444:4444 selenium/standalone-chrome

const parseURL = "https://russia.blsspainglobal.com/Global/account/login"
const maxTries = 10

func main() {
	// load config
	config, err := cfg.LoadConfig()
	if err != nil {
		log.Fatalln(err)
	}

	// chat api test
	//resp, err := api.GPT4oMiniRequest("", config.ChatApiKey)
	//if err != nil {
	//	log.Fatalln("chat gpt api error: ", err)
	//}
	//log.Println("chat api test msg: ", api.GetRespMsg(resp))
	//return

	// selenium connect
	seleniumService, err := service.NewSeleniumService(maxTries, parseURL)
	if err != nil {
		log.Print("web driver connection error: ", err)
		return
	}
	wd := seleniumService.Wd()
	defer wd.Quit()
	log.Println("web driver connected")

	if err = wd.MaximizeWindow(""); err != nil {
		log.Println("maximizing window error: ", err)
	}

	err = wd.Get(parseURL)
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

	//time.Sleep(time.Second * 10)
	log.Println(verifyBtn.Click())
	time.Sleep(time.Second * 5)

	err = seleniumService.ProcessCaptcha(wd)
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
