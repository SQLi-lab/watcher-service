package watcher

import (
	log "github.com/sirupsen/logrus"
	"watcher/internal/checker"
	"watcher/internal/scheduler"
)

// StartWatcher функция запускает watcher
func StartWatcher() {
	log.Info("Запуск Watcher-Service")

	// запуск горутины проверки статусов лаб
	go checker.StartChecker()

	// запуск сервиса отложенного удаления лаб
	scheduler.StartSchedulerServer()
}
