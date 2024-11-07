package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"
	"visasolution/internal/app"

	cfg "visasolution/internal/config"
	"visasolution/internal/service"
	"visasolution/internal/worker"
)

const (
	baseURL                 = "https://russia.blsspainglobal.com/"
	visaTypeVerificationURL = "Global/bls/VisaTypeVerification"
)

const (
	proxiesFilePath    = "./proxies.json"
	logFolder          = "logs/"
	logFilename        = "app.log"
	tmpFolder          = "tmp/"
	cookieFilename     = "cookies.json"
	screenshotFilename = "screenshot.png"
)

const (
	connectionMaxTries     = 10
	processCaptchaMaxTries = 5
)
const defaultMainLoopIntervalM = 30

var (
	mainLoopIntervalM        int
	availbilityNotifiedEmail string
)

func main() {
	parseFlags()

	logFile, err := setupLogger()
	if err != nil {
		log.Fatalln("Failed to setup logger:", err)
	}
	defer logFile.Close()
	log.Println("Logger initialized")

	ctx, cancel := setupSignalHandler()
	defer cancel()

	config, err := cfg.LoadConfig()
	if err != nil {
		log.Fatalln(err)
	}

	proxiesManager, err := worker.LoadProxies(proxiesFilePath)
	if err != nil {
		log.Println("Failed to load proxies from JSON:", err)
	}

	services := service.NewService(service.Deps{
		SeleniumURL:       config.SeleniumUrl,
		BaseURL:           baseURL,
		MaxTries:          connectionMaxTries,
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
			ScreenshotFilePath: path.Join(tmpFolder, screenshotFilename),
		},
	})

	workers := worker.NewWorker(services, worker.Deps{
		BaseURL:         baseURL,
		VisaTypeURL:     visaTypeVerificationURL,
		TmpFolder:       tmpFolder,
		CookieFile:      cookieFilename,
		NotifiedEmail:   availbilityNotifiedEmail,
		CaptchaMaxTries: processCaptchaMaxTries,
		ScreenshotFile:  screenshotFilename,
	})

	// TODO: refactor
	err = workers.MakePreparation()
	if err != nil {
		log.Fatalln("Make preparation error:", err)
	}

	err = services.Chat.ClientInitWithProxy(config.ProxyForeign)
	if err != nil {
		log.Fatalln("Chat client init error:", err)
	}
	log.Println("Chat API client initialized")

	err = services.Image.ClientInitWithProxy(proxiesManager.Current())
	if err != nil {
		log.Fatalln("Image client init error:", err)
	}
	log.Println("Image API client initialized")

	err = workers.ConnectGeneratedProxy(services.Selenium, proxiesManager.Current())
	if err != nil {
		log.Fatalln("Web driver connection error:", err)
	}
	log.Println("Web driver connected with proxy:", proxiesManager.Current().Host)
	defer services.Quit()
	defer workers.SaveCookies()

	app.RunMainLoop(ctx, app.MainLoopDeps{
		Workers:        workers,
		Services:       services,
		Config:         config,
		ProxiesManager: proxiesManager,
	}, mainLoopIntervalM)

	<-ctx.Done()
	log.Println("App stopped gracefully")
}

func setupSignalHandler() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		log.Println("Received interrupt signal, shutting down...")
		cancel()
	}()
	return ctx, cancel
}

func setupLogger() (*os.File, error) {
	if err := os.MkdirAll(logFolder, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create log folder: %v", err)
	}

	logFile, err := os.OpenFile(path.Join(logFolder, logFilename), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	return logFile, nil
}

// parseFlags парсит флаги командной строки
// -email - email для уведомлений о доступности. Обязательный
// -interval - интервал основного цикла в минутах
func parseFlags() {
	flag.IntVar(&mainLoopIntervalM, "interval", defaultMainLoopIntervalM, "Main loop interval in minutes")
	flag.StringVar(&availbilityNotifiedEmail, "email", "", "Email for availability notifications. Required")
	flag.Usage = func() {
		fmt.Println("Usage: bot [options]")
		flag.PrintDefaults()
	}
	flag.Parse()

	if availbilityNotifiedEmail == "" {
		log.Fatal("Email for availability notifications is required")
	}
}
