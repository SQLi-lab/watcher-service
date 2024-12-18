package scheduler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"
	"watcher/internal/postgres"
)

// RequestLab структура body запроса DELETE к deploy-service
type RequestLab struct {
	Name           string `json:"name"`
	UUID           string `json:"uuid"`
	ExpiredSeconds string `json:"expired_seconds"`
	DeploySecret   string `json:"deploy_secret,omitempty"`
}

// LabsManager планировщик, хранит глобальный мьютекс и мапу отложенных задач, переданных на удаление
type LabsManager struct {
	mu   sync.Mutex
	labs map[string]context.CancelFunc
}

// calculateTime функция вычисляет время остановки задачи
func calculateTime(data string) (time.Time, error) {
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return time.Time{}, fmt.Errorf("не удалось загрузить часовой пояс: %v", err)
	}

	layout := "2006-01-02 15:04:05.999999-07:00"
	targetTime, err := time.Parse(layout, data)

	if err != nil {
		return time.Time{}, fmt.Errorf("ошибка парсинга времени: %w", err)
	}
	targetTime = targetTime.In(location)

	return targetTime, nil
}

// newLabManager getter нового планировщика
func newLabManager() *LabsManager {
	return &LabsManager{
		labs: make(map[string]context.CancelFunc),
	}
}

// initFromDB метод инициализации задач LabManager при зарпуске с уже имеющимися задачами в системе
func (lm *LabsManager) initFromDB(db *sql.DB) {
	log.Infof("Запуск инициализации задач LanManager из БД")

	labs, err := postgres.GetActiveLabs(db)
	if err != nil {
		db = postgres.InitDatabase()
		labs, err = postgres.GetActiveLabs(db)
	}
	if err != nil {
		log.Errorf("Ошибка инициализации, невозможно получить список активных работ из БД: %v", err)
		return
	}

	if len(labs) == 0 {
		log.Infof("Задачи не были добавлены при инициализации")
		return
	}

	for _, lab := range labs {
		err := lm.createLabTsk(lab.UUID, lab.DateDeleted)
		if err != nil {
			log.Errorf("[ %s ] ошибка инициализации лабораторной")
		}
	}

	log.Infof("Конец инициализации работ")
}

// createLabTask метод добавления отложенной задачи в планировщик
func (lm *LabsManager) createLabTsk(uuid string, data string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if _, ok := lm.labs[uuid]; ok {
		log.Warnf("[ %s ] задача уже существует", uuid)
		return errors.New("задача уже существует")
	}

	targetTime, err := calculateTime(data)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	lm.labs[uuid] = cancel
	go lm.runTask(ctx, uuid, targetTime)
	log.Infof("[ %s ] Задача поставлена на удаление", uuid)

	return nil
}

// canselLabTask метод удаления задачи из планировщика
func (lm *LabsManager) deleteLabTask(uuid string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if _, ok := lm.labs[uuid]; ok {
		delete(lm.labs, uuid)
		log.Infof("[ %s ] задача удалена из LabManager", uuid)
	} else {
		log.Warnf("[ %s ] задача не найдена и не удалена из LabManager", uuid)
	}
}

// sendRequest метод отправки запроса на deploy-service для удаления лабы
func (lm *LabsManager) sendRequest(uuid string) {
	deployServiceURL := os.Getenv("DEPLOY_URL")
	if deployServiceURL == "" {
		deployServiceURL = "http://deploy-service:8001"
	}

	deployServiceURL = fmt.Sprintf("%s/api/v1/lab/delete", deployServiceURL)

	requestBody := RequestLab{
		Name:           uuid,
		UUID:           uuid,
		ExpiredSeconds: "10800",
		DeploySecret:   os.Getenv("DEPLOY_SECRET"),
	}

	data, err := json.Marshal(requestBody)
	if err != nil {
		log.Errorf("Ошибка формирования JSON: %v", err)
	}

	req, err := http.NewRequest("DELETE", deployServiceURL, bytes.NewBuffer(data))
	if err != nil {
		log.Errorf("Ошибка формирования запроса к deploy-service: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Ошибка выполнения запроса к deploy-service: %v", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		log.Errorf("Ошибка удаления лабораторной на стороне deploy-service: %s", string(body))
	}

	log.Infof("Лабораторная удалена, ответ deploy-service: %s", string(body))
}

// stopLabTask метод вызыватся при событии удалении задачи.
// Удаляет из планировщика и посылает запрос к deploy-service на удаление
func (lm *LabsManager) stopLabTask(uuid string) {
	lm.deleteLabTask(uuid)
	lm.sendRequest(uuid)
}

// runTask метод запуска ожидания удаления задачи
func (lm *LabsManager) runTask(ctx context.Context, uuid string, targetTime time.Time) {
	delay := time.Until(targetTime)

	if delay <= 0 {
		log.Warnf("[ %s ] время выполнения задачи уже прошло, запускаю удаление", uuid)
		cansel, ok := lm.labs[uuid]
		if !ok {
			log.Errorf("[ %s ] ошибка получения контекста удаления", uuid)
			return

		}
		cansel()
	}

	select {
	case <-time.After(delay):
		log.Infof("[ %s ] передача задачи на удаление", uuid)
		lm.stopLabTask(uuid)

	case <-ctx.Done():
		log.Infof("[ %s ] принудительное завершение", uuid)
		lm.stopLabTask(uuid)
	}
}
