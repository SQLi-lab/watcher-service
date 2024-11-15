package checker

import (
	"database/sql"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
	"sync"
	"time"
	"watcher/internal/docker"
	"watcher/internal/postgres"
)

func StartChecker() {
	var db = postgres.InitDatabase()

	log.Info("Запуск сервиса checker проверки состояний лабораторных")
	go checkerJob(db)

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
			go checkerJob(db)
		}
	}
}

func checkerJob(db *sql.DB) {
	log.Info("Запуск сервиса checker")

	// канал для горутины запросов к БД
	channelPostgres := make(chan struct {
		containersUUID []string
		err            error
	})

	// канал для горутины инспектора контенеров
	channelDocker := make(chan struct {
		containersJson []byte
		err            error
	})

	wg := sync.WaitGroup{}
	// параллельно запускаем запрос в БД и сбор инфы через Docker
	wg.Add(2)

	go func(db *sql.DB) {
		defer wg.Done()
		data, postgresErr := postgres.GetActiveContainers(db)
		channelPostgres <- struct {
			containersUUID []string
			err            error
		}{data, postgresErr}
	}(db)

	go func() {
		defer wg.Done()
		data, dockerErr := docker.GetDockerContainers()
		channelDocker <- struct {
			containersJson []byte
			err            error
		}{data, dockerErr}
	}()

	// собираем полученные данные из горутин
	resultPostgres := <-channelPostgres
	if resultPostgres.err != nil {
		return
	}

	resultDocker := <-channelDocker
	if resultDocker.err != nil {
		return
	}

	wg.Wait()
	log.Infof("Контейнеры получены от Docker")
	log.Debugf("Получены контейнеры: %v", string(resultDocker.containersJson))
	log.Infof("Список активных контейнеров получен")
	log.Debugf("Получены активные контейнеры: %v", resultPostgres.containersUUID)

	// TODO: проверка, что все активные лабы запущены и работают

	//err := postgres.SetErrorStatus(db, "318d020e-ee62-4235-b167-4587dcfc788b")
	//if err != nil {
	//	return
	//}
	//log.Warnf("Статус контейнера с ошибкой измененн")
}
