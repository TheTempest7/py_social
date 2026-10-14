// @title           Social Network API
// @version         1.0
// @description     API социальной сети
// @host            localhost:8080
// @BasePath        /

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/TheTempest7/social_network/app-go/internal/database"
	"github.com/TheTempest7/social_network/app-go/internal/handlers"
	"github.com/TheTempest7/social_network/app-go/internal/repository"
	"github.com/TheTempest7/social_network/app-go/internal/services"
	"github.com/joho/godotenv"

	_ "github.com/TheTempest7/social_network/app-go/docs"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {

	// Загружаем .env
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️ .env не найден, используем переменные окружения")
	}

	db, err := database.ConnectDB("data/social_app.db")
	if err != nil {
		fmt.Println("❌ Ошибка подключения к БД:", err)
		return
	}
	defer db.Close()
	fmt.Println("✅ Подключено к SQLite")

	if err := database.RunMigrations(db, "migrations"); err != nil {
		fmt.Println("❌ Ошибка миграций:", err)
		return
	}

	// 👇 Создаём слои
	repo := repository.NewUserRepository(db)
	userService := services.NewUserService(repo)
	authHandler := handlers.NewAuthHandler(userService)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Hello, World!",
		})
	})

	// 👇 Ручка /ping
	r.Get("/ping", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "pong",
		})
	})

	r.Get("/swagger/*", httpSwagger.WrapHandler)
	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/logout", authHandler.Logout)
	})

	fmt.Println("🚀 Сервер запущен на http://localhost:8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		fmt.Println("Ошибка сервера:", err)
	}
}
