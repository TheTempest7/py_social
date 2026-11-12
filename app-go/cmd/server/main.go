// @title           Social Network API
// @version         1.0
// @description     API социальной сети
// @host            localhost:8080
// @BasePath        /
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

package main

import (
	"fmt"
	"log/slog"
	"os"

	_ "github.com/TheTempest7/social_network/app-go/docs"
	"github.com/TheTempest7/social_network/app-go/internal/app"
	"github.com/TheTempest7/social_network/app-go/internal/config"

	"github.com/joho/godotenv"
)

func main() {

	// Загружаем .env
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️ .env не найден, используем переменные окружения")
	}

	config.Load()

	db, err := app.InitDB()
	if err != nil {
		os.Exit(1)
	}
	defer db.Close()
	storage, err := app.InitStorage()
	if err != nil {
		os.Exit(1)
	}

	broker, err := app.InitBroker() // 👈 Две переменные!
	if err != nil {
		slog.Error("❌ Ошибка RabbitMQ:", "error", err)
		os.Exit(1)
	}
	defer broker.Close()

	r := app.NewRouter(db, storage, broker)

	app.RunServer(r, config.AppConfig.ServerPort)
}
