package watcher

import (
	log "github.com/sirupsen/logrus"
	"os"
	"watcher/internal/checker"
	"watcher/internal/postgres"
	"watcher/internal/scheduler"
)

// StartWatcher функция запускает watcher
func StartWatcher() {
	log.Info("Запуск Watcher-Service")

	if os.Getenv("LOG_LEVEL") == "debug" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	// создание единого подключения для работы с БД
	var db = postgres.InitDatabase()

	// запуск горутины проверки статусов лаб
	go checker.StartChecker(db)

	// запуск сервиса отложенного удаления лаб
	scheduler.StartSchedulerServer(db)
}
