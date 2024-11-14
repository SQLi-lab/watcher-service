package watcher

import (
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type LabsManager struct {
	mu   sync.Mutex
	labs map[string]context.CancelFunc
}

func calculateTime(data string) (time.Time, error) {
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return time.Time{}, fmt.Errorf("не удалось загрузить часовой пояс: %v", err)
	}
	targetTime, err := time.Parse(time.RFC3339Nano, data)
	if err != nil {
		return time.Time{}, fmt.Errorf("ошибка парсинга времени: %w", err)
	}
	targetTime = targetTime.In(location)

	return targetTime, nil
}

func newLabManager() *LabsManager {
	return &LabsManager{
		labs: make(map[string]context.CancelFunc),
	}
}

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

func (lm *LabsManager) canselLabTsk(uuid string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if cancel, ok := lm.labs[uuid]; ok {
		cancel()
		delete(lm.labs, uuid)
		log.Infof("[ %s ] задача удалена", uuid)
	} else {
		log.Warnf("[ %s ] задача не найдена и не удалена", uuid)
		return errors.New("задача не найдена")
	}

	return nil
}

func (lm *LabsManager) runTask(ctx context.Context, uuid string, targetTime time.Time) {
	delay := time.Until(targetTime)

	if delay <= 0 {
		log.Warnf("[ %s ] время выполнения задачи уже прошло, запускаю удаление", uuid)
		// TODO: запрос на ручу удаления deploy-service
		return
	}

	select {
	case <-time.After(delay):
		log.Infof("[ %s ] передача задачи на удаление", uuid)
		// TODO: доделать запрос на ручку удаления deploy-service
	case <-ctx.Done():
		log.Infof("[ %s ] принудительное завершение", uuid)
	}
}
