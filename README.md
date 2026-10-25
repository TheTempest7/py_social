# Social Network — Go Backend

REST API для социальной сети на Go.

## 🚀 Быстрый старт

```bash
# 1. Перейти в папку бэкенда
cd ./app-go

# 2. Запустить Redis (в отдельном терминале)
docker run -d --name my-redis -p 6379:6379 redis:alpine

# 3. Запустить сервер
go run ./cmd/server

====================================================================================
Redis используется для черного списка (blacklist) отозванных JWT-токенов при логауте.

docker run -d --name my-redis -p 6379:6379 redis:alpine
docker ps
docker exec -it my-redis redis-cli
# Внутри:
keys *
exit
docker stop my-redis
docker rm my-redis
====================================================================================
Генерация Swagger:
cd ..
swag init -g app-go/cmd/server/main.go -o app-go/docs




====================================================================================
# Все тесты
go test ./...

# Подробный вывод
go test -v ./...

# Конкретный тест
go test -run TestRegisterUserRaceCondition ./app-go/internal/services/user/...

# С покрытием
go test -cover ./...

# С детектором гонок
go test -race ./...

====================================================================================

cd ./app-go
sqlite3 data/social_app.db

# Внутри (по одной команде):
.tables
.schema users
.mode column
.headers on
SELECT * FROM users;
.quit

====================================================================================

# Установка Go-зависимостей
go mod download

# Установка swag CLI (для генерации Swagger)
go install github.com/swaggo/swag/cmd/swag@latest
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc

cd ..
go mod tidy
go build ./...

====================================================================================

app-go/
├── cmd/server/main.go              # Точка входа (запуск сервера)
├── internal/
│   ├── config/                     # Конфигурация (.env)
│   ├── database/                   # Подключение к БД + миграции
│   ├── models/                     # Модели БД (User)
│   ├── dto/                        # Структуры запросов/ответов
│   ├── repository/                 # Работа с БД (SQL)
│   ├── services/                   # Бизнес-логика (user, profile)
│   ├── handlers/                   # HTTP-ручки
│   ├── middleware/                 # Мидлвари (JWT, CORS)
│   └── utils/                      # Утилиты (хелперы)
├── migrations/                     # SQL-миграции (golang-migrate)
├── docs/                           # Swagger (генерируется)
└── data/
    └── social_app.db               # SQLite БД