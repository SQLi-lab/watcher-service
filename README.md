# watcher-service
Сервис мониторинга состояний контейнеров + scheduler отложенного удаления

## Переменные окружения 
- `PORT`: порт сервера API; default `8002`
- `DEPLOY_SERVICE_URL`: адрес deploy-service; default http://deploy-service:8001
- `DEPLOY_SECRET`: секрет для общения с deploy-service

## Логика работы с backend и deploy-service
- При создании лабораторной шлется зарпос на deploy-service, при успешном развертывании от него 
шлется запрос на watcher, чтобы добавить задачу в scheduler на удаление.
- При удалении лабораторной сначала шлется запрос на watcher для удлаения задачи из scheduler, 
далее шлется зарпос на deploy-service, чтобы удалить окончательно лабораторную.
- При развертывании главное сначала создать задачу, а при удалении убрать из scheduler, 
чтобы он не отправил запрос на удаление уже удаленной задачи.