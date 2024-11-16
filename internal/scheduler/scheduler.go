package scheduler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

type Handler func(w http.ResponseWriter, r *http.Request) error

// JSONResponse структура ответа севрера в формате JSON
type JSONResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// LabResponse структура ответа на добавление/удаление лабы от deploy-service
type LabResponse struct {
	UUID string `json:"uuid"`
	Date string `json:"date"`
}

// ServeHTTP базовый метод обработки HTTP запроса
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		log.Errorf("Ошибка обработки запроса: %s", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(JSONResponse{
			Success: false,
			Message: err.Error(),
		})
	}
}

// LabAddHandler хэндлер для добавления отложенной задачи
func LabAddHandler(lm *LabsManager) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		var req LabResponse
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Некорректный формат данных", http.StatusBadRequest)
			return err
		}

		uuid := req.UUID
		date := req.Date

		if uuid == "" || date == "" {
			return errors.New("требуются параметры uuid и date")
		}

		err := lm.createLabTsk(uuid, date)
		if err != nil {
			return err
		}

		json.NewEncoder(w).Encode(JSONResponse{
			Success: true,
			Message: fmt.Sprintf("задача %s успешно добавлена", uuid),
		})
		return nil
	}
}

// LabCanselHandler хэндлер для удаления лабы из отложенной задачи
func LabCanselHandler(lm *LabsManager) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		var req LabResponse
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Некорректный формат данных", http.StatusBadRequest)
			return err
		}

		uuid := req.UUID

		if uuid == "" {
			return errors.New("требуются параметры uuid и date")
		}

		log.Infof("[ %s ] запрос на удаление задачи", uuid)

		cansel, ok := lm.labs[uuid]
		if !ok {
			// попытка запроса к deploy-service на удаление
			log.Warnf("[ %s ] задачи не существует, но пробую сделать запрос на удаление к deploy-service", uuid)
			lm.sendRequest(uuid)
			json.NewEncoder(w).Encode(JSONResponse{
				Success: true,
				Message: fmt.Sprintf("задачи %s не существует в LabManager", uuid),
			})
			return nil
		}

		// вызов отмены задачи
		cansel()

		json.NewEncoder(w).Encode(JSONResponse{
			Success: true,
			Message: fmt.Sprintf("задача %s успешно удалена", uuid),
		})
		return nil
	}
}

// StartSchedulerServer функция запуска API сервера
func StartSchedulerServer(db *sql.DB) {
	labManger := newLabManager()
	labManger.initFromDB(db)

	r := chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		r.Method("POST", "/add", LabAddHandler(labManger))
		r.Method("POST", "/cansel", LabCanselHandler(labManger))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}
	addr := fmt.Sprintf(":%s", port)

	log.Infof("Запуск API сервиса %s", addr)

	http.ListenAndServe(addr, r)
}
