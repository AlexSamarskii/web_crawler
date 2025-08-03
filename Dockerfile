FROM golang:1.22-alpine AS builder

# Устанавливаем необходимые зависимости
RUN apk add --no-cache git ca-certificates

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o crawler-cli ./

# Второй этап: минимальный образ для запуска
FROM alpine:latest AS final

# Устанавливаем сертификаты для HTTPS-запросов
RUN apk --no-cache add ca-certificates

# Устанавливаем рабочую директорию
WORKDIR /root/

# Копируем бинарник из builder-образа
COPY --from=builder /app/crawler-cli .
# Копируем конфигурационный файл
COPY --from=builder /app/config.yml .

# Команда по умолчанию
CMD ["./crawler-cli", "crawl", "--seed", "https://golang.org"]