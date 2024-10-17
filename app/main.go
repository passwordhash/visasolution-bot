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
}
