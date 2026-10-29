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
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/TheTempest7/social_network/app-go/docs"
	"github.com/TheTempest7/social_network/app-go/internal/auth"
	"github.com/TheTempest7/social_network/app-go/internal/config"
	"github.com/TheTempest7/social_network/app-go/internal/database"
	"github.com/TheTempest7/social_network/app-go/internal/handlers"
	appmw "github.com/TheTempest7/social_network/app-go/internal/middleware"
	"github.com/TheTempest7/social_network/app-go/internal/repository"
	"github.com/TheTempest7/social_network/app-go/internal/services/profile"
	"github.com/TheTempest7/social_network/app-go/internal/services/user"
	"github.com/TheTempest7/social_network/app-go/internal/storage"

	"github.com/joho/godotenv"

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

	config.Load()

	db, err := database.ConnectDB(config.AppConfig.DBPath)
	if err != nil {
		slog.Error("❌ Ошибка подключения к БД:", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	fmt.Println("✅ Подключено к SQLite")

	if err := database.RunMigrations(config.AppConfig.DBPath, config.AppConfig.MigrationsPath); err != nil {
		slog.Error("❌ Ошибка миграций:", "error", err)
		os.Exit(1)
	}

	storage, err := storage.NewStorage(
		config.AppConfig.MinioEndpoint,
		config.AppConfig.MinioAccessKey,
		config.AppConfig.MinioSecretKey,
		config.AppConfig.MinioBucket,
		config.AppConfig.MinioUseSSL,
	)
	if err != nil {
		slog.Error("❌ Ошибка MinIO:", "error", err)
		os.Exit(1)
	}
	fmt.Println("✅ Подключено к MinIO")

	// 👇 Создаём слои
	repo := repository.NewUserRepository(db)
	userService := user.NewUserService(repo)
	blacklist := auth.NewBlacklist(config.AppConfig.RedisAddr, config.AppConfig.RedisPassword)
	authHandler := handlers.NewAuthHandler(userService, blacklist)
	profileService := profile.NewProfileService(repo, storage)

	profileHandler := handlers.NewProfileHandler(profileService)

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
		// Публичные
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/logout", authHandler.Logout)

	})

	// 2. Группа пользователей (защищённая)
	r.Route("/api/users", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return appmw.RequireAuth(blacklist, next)
		})

		r.Get("/profile", profileHandler.Profile)
		r.Put("/profile", profileHandler.PutProfile)
		r.Put("/avatar", profileHandler.UploadAvatar)
	})
	fmt.Println("🚀 Сервер запущен на http://localhost:8080")

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second, // защита от Slowloris
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Запускаем сервер в отдельной горутине
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка сервера: %v", err)
		}
	}()

	// Ждём сигнал
	<-ctx.Done()

	// Даём 10 секунд на завершение текущих запросов
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Ошибка shutdown: %v", err)
	}
	fmt.Println("Сервер остановлен")
}
