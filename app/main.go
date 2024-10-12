package main

import (
	"log"
	cfg "visasolution/app/config"
	"visasolution/app/service"
	"visasolution/app/worker"
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

	services := service.NewService(service.Deps{
		MaxTries:          maxTries,
		BlsEmail:          config.BlsEmail,
		BlsPassword:       config.BlsPassword,
		ChatApiKey:        config.ChatApiKey,
		ImgurClientId:     config.ImgurClientId,
		ImgurClientSecret: config.ImgurClientSecret,
	})
	workers := worker.NewWorker(services, parseURL)

	// TODO: client imgur

	// Chat client init
	err = services.Chat.ClientInitWithProxy(config.ProxyRowForeign)
	if err != nil {
		log.Fatalln("chat client init error:", err)
	}
	log.Println("chat api client inited")

	// Selenium connect
	err = services.Selenium.Connect("")
	if err != nil {
		log.Println("web driver connection error: ", err)
		return
	}
	defer services.Quit()
	log.Println("web driver connected")

	// Run worker
	err = workers.Run()
	if err != nil {
		log.Println("worker run error:", err)
		return
	}

	//// form
	//emailInput, err := wd.FindElement(selenium.ByXPATH, "//*[@id=\"UserId6\"]")
	//if err != nil {
	//	log.Println("cannot get email input element: ", err)
	//	return
	//}
	//
	//passwordInput, err := wd.FindElement(selenium.ByXPATH, "//*[@id=\"Password2\"]")
	//if err != nil {
	//	log.Println("cannot get password input element: ", err)
	//	return
	//}
	//
	//verifyBtn, err := wd.FindElement(selenium.ByXPATH, "//*[@id=\"btnVerify\"]")
	//if err != nil {
	//	log.Println("cannot get verify btn: ", err)
	//	return
	//}
	//
	////time.Sleep(time.Second * 10)
	//log.Println(verifyBtn.Click())
	//time.Sleep(time.Second * 5)
	//
	//err = seleniumService.ProcessCaptcha(wd)
	//if err != nil {
	//	log.Println("error while proccessing captcha: ", err)
	//	return
	//}
	//
	//time.Sleep(time.Second * 10)
	//
	//// FIXME: not iteractable
	//err = emailInput.SendKeys(config.BlsEmail)
	//if err != nil {
	//	log.Println("cannot send key in email input: ", err)
	//	return
	//}
	//
	//// FIXME: not iteractable
	//err = passwordInput.SendKeys(config.BlsPassword)
	//if err != nil {
	//	log.Println("cannot send key in password input: ", err)
	//	return
	//}
	//
	//fmt.Println("finish")
}
