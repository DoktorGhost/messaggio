FROM golang:1.22-alpine

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем все файлы проекта в рабочую директорию контейнера
COPY . .

# Сборка Go-приложения
RUN go build -o client ./CMD/client/main.go

# Запуск приложения
CMD ["./client"]