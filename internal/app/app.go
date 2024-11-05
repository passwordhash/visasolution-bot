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
			// Выполнение основной логики воркера
			runErr := deps.Workers.Run()

			// Проверяем, произошла ли ошибка TooManyRequestsErr
			if shouldRestart := handleRunError(runErr, deps); shouldRestart {
				log.Println("Immediate restart of the main loop due to TooManyRequestsErr...")
				continue // Немедленный перезапуск цикла
			}

			// Ждем указанный интервал перед следующей итерацией
			select {
			case <-ctx.Done():
				log.Println("Context canceled, stopping main loop...")
				return
			case <-time.After(time.Duration(interval) * time.Minute):
				// Продолжаем к следующей итерации после ожидания
			}
		}
	}
}

// handleRunError обработка ошибок в основном цикле
func handleRunError(err error, deps MainLoopDeps) bool {
	if err == nil {
		return false
	}

	var banErr worker.TooManyRequestsErr
	if errors.As(err, &banErr) {
		log.Println(banErr)
		log.Println("Trying to reconnect with another proxy...")

		err := deps.Services.Selenium.Quit()
		if err != nil {
			log.Println("Web driver quit error:", err)
		}

		err = deps.Workers.ConnectWithGeneratedProxy(deps.Services.Selenium, deps.ProxiesManager.Next())
		if err != nil {
			log.Println("Web driver reconnect error:", err)
			return false
		}

		log.Println("Web driver reconnected with new proxy:", deps.ProxiesManager.Current().Host)
		return true
	}

	log.Println("Main loop error:", err)
	return false
}
