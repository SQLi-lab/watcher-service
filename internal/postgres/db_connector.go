package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"os"
)

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
