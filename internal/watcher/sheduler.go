package watcher

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
		uuid := r.URL.Query().Get("uuid")
		date := r.URL.Query().Get("date")

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
		uuid := r.URL.Query().Get("uuid")

		if uuid == "" {
			return errors.New("требуются параметры uuid и date")
		}

		log.Infof("[ %s ] запрос на удаление задачи", uuid)

		err := lm.canselLabTask(uuid)
		if err != nil {
			return err
		}

		json.NewEncoder(w).Encode(JSONResponse{
			Success: true,
			Message: fmt.Sprintf("задача %s успешно удалена", uuid),
		})
		return nil
	}
}

// startSchedulerServer функция запуска API сервера
func startSchedulerServer() {
	labManger := newLabManager()

	r := chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		r.Method("GET", "/add", LabAddHandler(labManger))
		r.Method("GET", "/cansel", LabCanselHandler(labManger))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}
	addr := fmt.Sprintf("127.0.0.1:%s", port)

	log.Infof("Запуск API сервиса %s", addr)

	http.ListenAndServe(addr, r)
}
