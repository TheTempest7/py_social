# Social Network — Go Backend

## Быстрый старт

```bash
cd ~/social_network/app-go
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

cd ~/social_network
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
cd ~/social_network/app-go

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




## Управление зависимостями (go.mod)

### Что такое `go.mod` и `go.sum`

- **`go.mod`** — список **прямых** и **косвенных** зависимостей проекта
- **`go.sum`** — хеши пакетов для проверки целостности (не редактируется вручную!)

### Что делает `go mod tidy`

**Проходит по всем `.go` файлам, читает импорты и синхронизирует `go.mod`:**

| Что делает | Пример |
|------------|--------|
| **Добавляет недостающие пакеты** | Импорт есть, в `go.mod` нет → добавит |
| **Удаляет лишние пакеты** | В `go.mod` есть, в коде не используется → удалит |
| **Помечает `// indirect`** | Пакет не импортируется напрямую, но нужен другой зависимости |
| **Обновляет `go.sum`** | Синхронизирует хеши |

**⚠️ `go mod tidy` НЕ трогает `.go` файлы — только `go.mod` и `go.sum`.**

### Как запускать

```bash
# 1. Перейти в папку модуля (где лежит go.mod)
cd ~/social_network

# 2. Почистить зависимости
go mod tidy

# 3. Проверить, что всё компилируется
go build ./...






**Имя файла:** `X_test.go` — рядом с `X.go`.

**Имя теста:** `TestЧто_Когда_Ожидание` — `TestRegisterUser_RaceCondition`.

**Пакет:** `package services_test` — снаружи (тестируем публичный API).

### Запуск

```bash
# Все тесты
go test ./...

# Подробный вывод
go test -v ./...

# Конкретный тест
go test -run TestRegisterUser_RaceCondition ./app-go/internal/services/...

# С покрытием
go test -cover ./...

# С детектором гонок
go test -race ./...