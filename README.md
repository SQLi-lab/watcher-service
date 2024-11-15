# watcher-service
Сервис мониторинга состояний контейнеров + scheduler отложенного удаления


## Логика работы _scheduler_ с _backend_ и _deploy-service_
- Запускается API-сервер на добавление и удаление лабораторной
- При создании лабораторной шлется зарпос на deploy-service, при успешном развертывании от него 
шлется запрос на watcher, чтобы добавить задачу в scheduler на удаление.
- При удалении лабораторной сначала шлется запрос на watcher для удлаения задачи из scheduler, 
далее шлется зарпос на deploy-service, чтобы удалить окончательно лабораторную.
- При развертывании главное сначала создать задачу, а при удалении убрать из scheduler, 
чтобы он не отправил запрос на удаление уже удаленной задачи.


## Логика работы _checker_
- Сервис запускается с таймаутом `CHECKER_TIMEOUT`
- Делает запрос в БД бэкенда, получает список активных лаб
- Собирает инфорамцию  статусах контейнеров из docker 
- При ошибке рабоыт контейнера меняет статус в БД бэкенда


## Переменные окружения
- `PORT`: порт сервера API; default `8002`
- `DEPLOY_SERVICE_URL`: адрес deploy-service; default http://deploy-service:8001
- `DEPLOY_SECRET`: секрет для общения с deploy-service
- `CHECKER_TIMEOUT`: тамаут работы сервиса checker в секундах; default `60`
- `LOG_LEVEL`: уровень логирования; default `info` # debug
- `POSTGRES_HOST`: хост БД postgres; default `localhost`
- `POSTGRES_PORT`: порт БД postgres; default `5432`
- `POSTGRES_DB`: имя БД postgres; default `sqli_lab`
- `POSTGRES_USER`: имя пользователя БД postgres; default `sqli_user`
- `POSTGRES_PASS`: пароль пользователя БД postgres; default `sqli_pass`
