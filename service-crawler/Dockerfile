# --- Stage 1: Build ---
FROM golang:1.21-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Сначала копируем только файлы модулей для кеширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Теперь копируем весь остальной код
COPY . .

# Собираем проект.
# CGO_ENABLED=0 нужен для статической линковки (чтобы работало в пустом образе)
RUN CGO_ENABLED=0 GOOS=linux go build -o crawler ./cmd/main.go

# --- Stage 2: Final ---
FROM alpine:latest

# Устанавливаем сертификаты (обязательно для краулера, чтобы скачивать HTTPS сайты)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарник из билдера
COPY --from=builder /app/crawler .
# Копируем .env если он нужен внутри контейнера (хотя лучше передавать через docker-compose)
COPY --from=builder /app/.env .

# Запуск
ENTRYPOINT ["./crawler"]
