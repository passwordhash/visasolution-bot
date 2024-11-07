package app

import (
	"context"
	"errors"
	"log"
	"time"
	"visasolution/internal/config"
	"visasolution/internal/service"
	"visasolution/internal/worker"
)

// MainLoopDeps структура для зависимостей основного цикла
type MainLoopDeps struct {
	Workers        *worker.Worker
	Services       *service.Service
	Config         *config.Config
	ProxiesManager *config.ProxiesManager
}

// RunMainLoop основной цикл приложения
func RunMainLoop(ctx context.Context, deps MainLoopDeps, interval int) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Context canceled, stopping main loop...")
			return
		default:
			runErr := deps.Workers.Run()

			shouldRestart := handleRunError(runErr, deps)
			if shouldRestart {
				log.Println("Restarting main loop...")
				continue
			}

			log.Println("Waiting for", interval, "minutes...")

			select {
			case <-ctx.Done():
				log.Println("Context canceled, stopping main loop...")
				return
			case <-time.After(time.Duration(interval) * time.Minute):
			}
		}
	}
}

// TODO: возврат еще и ошибки (обработать случай, когда не удалось переподключиться к Selenium)
// handleRunError обработка ошибок в основном цикле
// Возвращает true, если нужно перезапустить цикл, при этом перезапускает веб-драйвер с новым прокси
// Возможна ситуация, когда не удалось переподключиться к Selenium
func handleRunError(err error, deps MainLoopDeps) bool {
	if err == nil {
		return false
	}

	var connectErr worker.WDConnectError
	if errors.As(err, &connectErr) {
		err = deps.Workers.ConnectSameProxy(deps.Services.Selenium)
		if err != nil {
			log.Println("Web driver reconnect error:", err)
			return false
		}

		log.Println("Web driver reconnected with the same proxy:", deps.ProxiesManager.Current().Host)
		return true
	}

	var banErr worker.TooManyRequestsErr
	if errors.As(err, &banErr) {
		log.Println(banErr)
		log.Println("Trying to reconnect with another proxy...")

		err := deps.Services.Selenium.Quit()
		if err != nil {
			log.Println("Web driver quit error:", err)
		}

		newProxie := deps.ProxiesManager.Next()
		err = deps.Workers.ConnectGeneratedProxy(deps.Services.Selenium, newProxie)
		if err != nil {
			log.Println("Web driver reconnect error:", err)
			return false
		}

		log.Println("Web driver reconnected with new proxy:", newProxie.Host)
		return true
	}

	log.Println("Main loop error:", err)
	return false
}
