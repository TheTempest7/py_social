FROM golang:1.27-alpine AS builder

# 👇 Заходим в папку, которая совпадает с путем из go.mod
WORKDIR /src/app-go

# Копируем go.mod и go.sum (из корня контекста)
COPY go.mod go.sum /src/
WORKDIR /src
RUN go mod download

# Возвращаемся в папку app-go и копируем туда код
WORKDIR /src/app-go
COPY app-go/ .

# Собираем (теперь пути будут совпадать)
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

# Финальный образ
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /src/app-go/migrations ./migrations
EXPOSE 8080
CMD ["./server"]