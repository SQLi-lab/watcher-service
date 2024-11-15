package checker

import (
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
	"time"
)

func StartChecker() {
	log.Info("Запуск сервиса checker проверки состояний лабораторных")
	go checkerJob()

	timeoutStr := os.Getenv("CHECKER_TIMEOUT")
	if timeoutStr == "" {
		timeoutStr = "60"
	}
	timeout, _ := strconv.Atoi(timeoutStr)
	log.Infof("Сервис checker работает с таймаутом: %s", timeoutStr)

	ticker := time.NewTicker(time.Duration(timeout) * time.Second)

	for {
		select {
		case <-ticker.C:
			go checkerJob()
		}
	}
}

func checkerJob() {
	log.Info("Запуск сервиса checker")
	// TODO:  запрос к БД на полученеи активных лаб -> json

	// TODO: получение списка контейнеров -> json

	// TODO: проверка, что все активные лабы запущены и работают

	// TODO: sleep TIMEOUT
}
