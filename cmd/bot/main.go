package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	cfg "visasolution/internal/config"
	"visasolution/internal/service"
	"visasolution/internal/worker"
)

const (
	baseURL                 = "https://russia.blsspainglobal.com/"
	visaTypeVerificationURL = "Global/bls/VisaTypeVerification"
)

const (
	tmpFolder      = "tmp/"
	cookieFile     = "cookies.json"
	screenshotFile = "screenshot.png"
)

const (
	maxTries               = 10
	processCaptchaMaxTries = 3
)

const mainLoopIntervalM = 10

const availbilityNotifiedEmail = "iam@it-yaroslav.ru"

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	go handleDoneSigs(cancel)

	config, err := cfg.LoadConfig()
	if err != nil {
		log.Fatalln(err)
	}

	services := service.NewService(service.Deps{
		BaseURL:           baseURL,
		MaxTries:          maxTries,
		BlsEmail:          config.BlsEmail,
		BlsPassword:       config.BlsPassword,
		ChatApiKey:        config.ChatApiKey,
		ImgurClientId:     config.ImgurClientId,
		ImgurClientSecret: config.ImgurClientSecret,
		EmailDeps: service.EmailDeps{
			Host:               config.SmtpHost,
			Port:               config.SmtpPort,
			Username:           config.SmtpUsername,
			Password:           config.Password,
			ScreenshotFilePath: tmpFolder + screenshotFile,
		},
	})

	workers := worker.NewWorker(services, worker.Deps{
		BaseURL:         baseURL,
		VisaTypeURL:     visaTypeVerificationURL,
		TmpFolder:       tmpFolder,
		CookieFile:      cookieFile,
		NotifiedEmail:   availbilityNotifiedEmail,
		CaptchaMaxTries: processCaptchaMaxTries,
		ScreenshotFile:  screenshotFile,
	})

	// Make preparatin
	err = workers.MakePreparation()
	if err != nil {
		log.Fatalln("Make preparation error:", err)
	}

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
	err = services.Selenium.ConnectWithProxy(config.SeleniumUrl, extensionPath)
	if err != nil {
		log.Println("Web driver connection error: ", err)
		return
	}
	log.Println("Web driver connected")
	defer services.Quit()
	defer workers.SaveCookies()

	// Main loop
	startPeriodicTask(ctx, mainLoopIntervalM*time.Minute, func() {
		err := workers.Run()
		if err != nil {
			log.Println("Main loop error:", err)
		}
		log.Println("Waiting for the next iteration ...")
	})
}

func startPeriodicTask(ctx context.Context, interval time.Duration, f func()) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			f()
		}
	}
}

func handleDoneSigs(cancel context.CancelFunc) {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigs
	fmt.Println("Signal received:", sig)

	cancel()
}
