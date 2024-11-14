package watcher

import (
	log "github.com/sirupsen/logrus"
)

func StartWatcher() {
	log.Info("Запуск Watcher-Service")

	// запуск горутины проверки статусов лаб
	go startChecker()

	startSchedulerServer()
}
