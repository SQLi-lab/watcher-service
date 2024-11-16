# watcher-service
Сервис мониторинга состояний контейнеров (_checker_) + сервис отложенного удаления лабораторных (_scheduler_).
Поскольку сервис рабоатет с `docker.sock` хоста, то он запускается от root и не должен открывать порты наружу, все общение с сервисом происходит внутри сети docker.


## Логика работы _scheduler_ 
Watcher запускает API-сервер на добавление и удаление лабораторной: 
```api
POST /api/v1/add 
POST /api/v1/cansel
```


- При создании лабораторной из backend шлется зарпос на deploy-service
- При успешном развертывании deploy-service шлет запрос на watcher, чтобы добавить задачу в scheduler на удаление


- При удалении лабораторной из backend шлется запрос на watcher для удлаения задачи из scheduler
- Далее шлется зарпос на deploy-service, чтобы удалить окончательно лабораторную


## Логика работы _checker_
Запускается горутина, которая через определенные интервалы запускает `checkerJob` проверки состояний контейнеров.
Также при старте запускается инициализация, которая собиарет из БД все активные лабы и их даты остановок, обновляет 
мапу LabManager и ставит задачи на удаление.
- Сервис запускается с таймаутом `CHECKER_TIMEOUT`
- Из БД бэкенда получает список активных лаб
- Собирает инфорамцию  статусах контейнеров из docker 
- При ошибке рабоыт контейнера меняет статус в БД бэкенда и отпарвляет запрос на себя (watcher) `POST /api/v1/cansel` для удаления лабораторной


## Переменные окружения
- `WATCHER_PORT`: порт сервера API; default `8002`
- `DEPLOY_SERVICE_URL`: адрес deploy-service; default http://deploy-service:8001
- `DEPLOY_SECRET`: секрет для общения с deploy-service
- `CHECKER_TIMEOUT`: тамаут работы сервиса checker в секундах; default `60`
- `LOG_LEVEL`: уровень логирования; default `info` # debug
- `POSTGRES_HOST`: хост БД postgres; default `localhost`
- `POSTGRES_PORT`: порт БД postgres; default `5432`
- `POSTGRES_DB`: имя БД postgres; default `sqli_lab`
- `POSTGRES_USER`: имя пользователя БД postgres; default `sqli_user`
- `POSTGRES_PASS`: пароль пользователя БД postgres; default `sqli_pass`


## Запуск Watcher 
1. Задать переменные окружения (`.env.example`), при локлаьном запуске сделать экспорт
2. Запуск в docker: `docker compose -f docker/docker-compose.yml up --build`
3. Запуск локально: `make build && ./watcher`

---
_Go 1.23.0_