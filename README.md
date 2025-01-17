<p align="left">
      <img src="https://i.ibb.co/cYzQsPG/logoza-ru.png" alt="Project Logo" width="726">
</p>

# Микросервис обработки сообщений
Задание выглядит так:

*Разработать микросервис на Go, который будет принимать сообщения через HTTP API,
сохранять их в PostgreSQL, а затем отправлять в Kafka для дальнейшей обработки.
Обработанные сообщения должны помечаться. Сервис должен также предоставлять
API для получения статистики по обработанным сообщениям.*

Непонятные моменты: кто читает сообщения из Kafka и кто их обрабатывает?

План:
1. Приложение А получает сообщения через HTTP API, сохраняет в БД и отправляет в Kafka.
2. Приложение Б ждет сообщений от Kafka, выполняет условную обработку, отправляет в Kafka сигнал, что сообщение обработано.
3. Приложение А ждет ответ от Kafka, получив ответ обновляет статус сообщения в БД.


### Настройка переменных окружения

```yaml
DB_HOST=db #хост бд
DB_PORT=5432    #порт бд
DB_NAME=admin #имя бд
DB_LOGIN=admin #логин для подключения к бд
DB_PASS=admin #пароль для подключения к бд
SERVER_HOST=0.0.0.0 #хост для работы сервера
SERVER_PORT=8080 #порт для работы сервера
SWAGGER_HOST=localhost
SWAGGER_PORT=8080
```
## Запуск контейнера
Собираем образ и поднимаем контейнер:

```golang
docker compose up
```

Поднимается 5 контейнеров zookeeper, kafka, postgres, приложение А, приложение Б.

## Тестирование
Если переменные окружения не менялись и приложение развернуто в локальной среде, можем пртестировать в Postman и в Swagger.

### Swagger
Открываем любой браузер и переходим по адресу http://localhost:8080/swagger/

<p align="left">
      <img src="https://i.ibb.co/47hnYs5/image.jpg" alt="Project Logo" width="726">
</p>

У нас есть два ендпоинта: /message и /stats
/message принимает в теле сообщение вида:
```golang
{
  "content": "string"
}
```
/stats выводит статистику по сообщениям в таком виде:
```golang
{
  "pending": 0,
  "processed": 0,
  "total": 0
}
```
где pending - ожидает обработки, processed - сообщение обработано, total - всего сообщений. 

### Postman
Открываем Postman, выбираем метод POST, строка запроса http://localhost:8080/message, в теле запроса вставляем сообщение
```golang
{
    "content": "Делать нечего в селе, мы сидим на веселе"
}
```
В случае успешного ответа получаем код 200.
Сообщение записывается в БД и отправляется в Kafka. Второе приложение получает сообщение и выполняет "обработку" в течении 5 секунд, после чего отправляет в Kafka ID данного сообщения. Первое приложение получает от Kafka ID сообщения и меняет в базе данных статус этого сообщения.

Выбираем метод GET, строка запроса http://localhost:8080/stats.
Если мы выполним данный запрос до того, как второй сервис успеет обработать сообщение, то увидим, что у нас есть необработанные сообщения:
```golang
{
    "total": 1,
    "processed": 0,
    "pending": 1
}
```


Когда обработка будет выполнена, получим другой результат:
```golang
{
    "total": 1,
    "processed": 1,
    "pending": 0
}
```
