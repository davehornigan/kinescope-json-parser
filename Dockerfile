# Stage 1: Build
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Копируем исходники
COPY . .

# Собираем статический бинарник
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -o /kinescope-json-parser

# Stage 2: Minimal runtime
FROM scratch

# Копируем только бинарник
COPY --from=builder /kinescope-json-parser /kinescope-json-parser

# Слушаем 8080 порт
EXPOSE 8080

# Стартуем приложение
ENTRYPOINT ["/kinescope-json-parser"]