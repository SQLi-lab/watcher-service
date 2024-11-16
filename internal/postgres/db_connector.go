package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

// ActiveLab структура активной абораторной, полученной из БД, для инициализации
type ActiveLab struct {
	UUID        string
	DateDeleted string
}

// InitDatabase функция инициализирует БД и возвращает экземпляр БД
func InitDatabase() *sql.DB {
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "127.0.0.2"
	}

	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		port = "5432"
	}

	user := os.Getenv("POSTGRES_USER")
	if user == "" {
		user = "sqli_user"
	}

	password := os.Getenv("POSTGRES_PASS")
	if password == "" {
		password = "sqli_pass"
	}

	database := os.Getenv("POSTGRES_DB")
	if database == "" {
		database = "sqli_lab"
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, database)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД Postgres: %v", err)
		return nil
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		log.Fatalf("Ошибка подключения к БД Postgres: %v", err)
		return nil
	}

	log.Infof("Успешное подключение к БД Postgres")

	return db
}

// GetActiveContainers функция получения активных контенйеров из БД, которым осталось работать больше 1 минуты
func GetActiveContainers(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT uuid FROM sqli_api_lab WHERE status = 'Выполняется' AND NOW() < date_started + INTERVAL '1 second' * expired_seconds - INTERVAL '1 minute'")
	if err != nil {
		log.Errorf("Ошибка получения данных из БД: %v", err)
		return nil, err
	}
	defer rows.Close()

	var containers []string
	for rows.Next() {
		var container string
		if err := rows.Scan(&container); err != nil {
			log.Errorf("Ошибка чтения данных от БД: %v", err)
			return nil, err
		}
		containers = append(containers, container)
	}

	return containers, nil
}

// GetActiveLabs функция получает список всех активных лаб со статусом "Выполняется" и высчитывает дату остановки
// возвращает массив структура ActiveLab с данными о времени удаления
func GetActiveLabs(db *sql.DB) ([]ActiveLab, error) {
	rows, err := db.Query("SELECT uuid, date_started, expired_seconds FROM sqli_api_lab WHERE status = 'Выполняется'")
	if err != nil {
		log.Errorf("Ошибка получения данных из БД: %v", err)
		return nil, err
	}

	defer rows.Close()

	var labs []ActiveLab

	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Printf("Ошибка загрузки часового пояса: %v", err)
		return nil, err
	}

	for rows.Next() {
		var uuid string
		var dateStarted time.Time
		var expiredSeconds int

		if err := rows.Scan(&uuid, &dateStarted, &expiredSeconds); err != nil {
			log.Errorf("Ошибка чтения данных от БД: %v", err)
			return nil, err
		}

		deleteTime := dateStarted.Add(time.Duration(expiredSeconds) * time.Second).In(location)

		labs = append(labs, ActiveLab{UUID: uuid, DateDeleted: deleteTime.Format("2006-01-02 15:04:05.999999-07:00")})
	}

	return labs, nil
}

// SetErrorStatus функция смены статуса контейнера с ошибкой
func SetErrorStatus(db *sql.DB, uuid string) error {
	updateStmt := `UPDATE sqli_api_lab SET "status"=$1, "error_log"=$2 where "uuid"=$3`
	_, err := db.Exec(updateStmt, "Ошибка выполнения", "Контейнер был остановлен, возникла ошибка", uuid)
	if err != nil {
		log.Errorf("Ошибка смены статуса поломанного контейнера: %v", err)
		return err
	}

	return nil
}
