package checker

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
	"watcher/internal/docker"
	"watcher/internal/postgres"
)

// WatcherCanselRequest структура запроса на удлаление к Watcher
type WatcherCanselRequest struct {
	UUID string `json:"uuid"`
}

// StartChecker функция запускает запуск проверки лабораторных
func StartChecker(db *sql.DB) {

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

// checkerJob горутина проверки лабораторных
func checkerJob(db *sql.DB) {
	// канал для горутины запросов к БД
	channelPostgres := make(chan struct {
		containersUUID []string
		err            error
	})
	// канал для горутины инспектора контенеров
	channelDocker := make(chan struct {
		containersMap map[string]string
		err           error
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
			containersMap map[string]string
			err           error
		}{data, dockerErr}
	}()

	// собираем полученные данные из горутин
	resultPostgres, resultDocker := <-channelPostgres, <-channelDocker
	if resultPostgres.err != nil || resultDocker.err != nil {
		return
	}

	wg.Wait()
	log.Infof("Получены контейнеры от Docker")
	log.Debugf("Получены контейнеры: %v", resultDocker.containersMap)
	log.Infof("Получены контейнеры от Backernd Postgres")
	log.Debugf("Получены активные контейнеры: %v", resultPostgres.containersUUID)

	wg.Add(len(resultPostgres.containersUUID))

	// проверка всех контейнеров в горутинах
	for _, containerUUID := range resultPostgres.containersUUID {
		go func(uuid string) {
			defer wg.Done()

			status := resultDocker.containersMap[uuid]
			if status == "running" {
				return
			}

			err := postgres.SetErrorStatus(db, uuid)
			if err != nil {
				return
			}
			log.Warnf("Статус контейнера с ошибкой %s изменен", uuid)

			err = cancelLab(uuid)
			if err != nil {
				return
			}
		}(containerUUID)
	}

	wg.Wait()

	log.Infof("Проверка контейнеров завершена")
	log.Infof("Ожидание следующего запуска...")
}

// canselLab функция отправляет запрос на сам watcher API для отмены лабораторнйо из scheduler
func cancelLab(uuid string) error {
	data := WatcherCanselRequest{
		UUID: uuid,
	}
	dataJson, err := json.Marshal(data)
	dataBuf := bytes.NewBuffer(dataJson)

	resp, err := http.Post("http://127.0.0.1:8002/api/v1/cansel", "application/json", dataBuf)
	if err != nil {
		log.Errorf("Ошибка формирования запроса: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Errorf("Ошибка выполнения запроса, код: %d", resp.StatusCode)
		return errors.New("ошибка выполенния запроса")
	}
	log.Infof("Запрос на отмену контейнера %s отправлен успешно", uuid)
	return nil
}
