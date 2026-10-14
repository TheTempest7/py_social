# Social Network — Go Backend

## Быстрый старт

```bash
cd ~/job/go/social_network/app-go
go run ./cmd/server



Установка зависимостей


go get github.com/go-chi/chi/v5
go get github.com/go-chi/cors
go get modernc.org/sqlite
go get golang.org/x/crypto/bcrypt
go get github.com/golang-jwt/jwt/v5
go get github.com/swaggo/http-swagger
go get github.com/swaggo/files



Установка swag CLI
go install github.com/swaggo/swag/cmd/swag@latest
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc


Генерация Swagger docs

cd ~/job/go/social_network
swag init -g app-go/cmd/server/main.go -o app-go/docs




Структура проекта

app-go/
├── cmd/server/main.go              ← точка входа
├── internal/
│   ├── config/                     ← конфиг
│   ├── database/                   ← подключение к БД + миграции
│   ├── models/                     ← модели (User)
│   ├── dto/                        ← структуры запросов/ответов
│   ├── repository/                 ← работа с БД (SQL)
│   ├── services/                   ← бизнес-логика
│   ├── handlers/                   ← HTTP-ручки
│   └── middleware/                 ← мидлвари
├── migrations/                     ← SQL-миграции
├── docs/                           ← Swagger (генерируется)
└── data/
    └── social_app.db               ← SQLite БД




    ## Работа с БД (SQLite)

# 1. Перейти в app-go
cd ~/job/go/social_network/app-go

# 2. Проверить БД есть
ls -la data/social_app.db

# 3. Открыть
sqlite3 data/social_app.db

# Внутри — по одной команде:
.tables
.schema users
.mode column
.headers on
SELECT * FROM users;

# Выйти
.quit