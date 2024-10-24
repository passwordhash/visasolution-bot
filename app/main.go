package main

import (
	"log"
	cfg "visasolution/app/config"
	"visasolution/app/service"
	"visasolution/app/worker"
)

//docker run --rm -p=4444:4444 selenium/standalone-chrome
//docker run --rm -p=4444:4444 --shm-size=2g -v /Users/yaroslav/code/projects/visasolution/volumes:/home/seluser/Downloads selenium/standalone-chrome

// const parseURL = "https://russia.blsspainglobal.com/Global/Bls/VisaTypeVerification"
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
		log.Fatalln("Chat client init error:", err)
	}
	log.Println("Chat api client inited")

	// Generate proxy auth extension
	extensionPath, err := workers.GenerateProxyAuthExtension(config.ProxyRow)
	if err != nil {
		log.Println("Generate proxy auth extension error:", err)
	}

	// Selenium connect
	err = services.Selenium.ConnectWithProxy("", extensionPath)
	if err != nil {
		log.Println("Web driver connection error: ", err)
		return
	}
	defer services.Quit()
	defer workers.SaveCookies()
	log.Println("Web driver connected")

	// Load cookies
	err = workers.LoadCookies()
	if err != nil {
		log.Println("Cookies load error:", err)
	}
	log.Println("Cookies loaded")

	// Run worker
	err = workers.Run()
	if err != nil {
		log.Println("Worker run error:", err)
		return
	}
}
