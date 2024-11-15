package scheduler

import (
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

type LabAddRequest struct {
	UUID string `json:"uuid"`
	Date string `json:"date"`
}

type LabDeleteRequest struct {
	UUID string `json:"uuid"`
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
		var req LabAddRequest
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
		var req LabAddRequest
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
			message := fmt.Sprintf("задачи %s не существует", uuid)
			return errors.New(message)
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
func StartSchedulerServer() {
	labManger := newLabManager()

	// TODO: запрос к бд на дамп активных лаб в labManager

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
