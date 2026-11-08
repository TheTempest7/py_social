package app

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/TheTempest7/social_network/app-go/internal/auth"
	"github.com/TheTempest7/social_network/app-go/internal/config"
	"github.com/TheTempest7/social_network/app-go/internal/handlers"
	appmw "github.com/TheTempest7/social_network/app-go/internal/middleware"
	"github.com/TheTempest7/social_network/app-go/internal/repository"
	"github.com/TheTempest7/social_network/app-go/internal/services/profile"
	"github.com/TheTempest7/social_network/app-go/internal/services/user"
	"github.com/TheTempest7/social_network/app-go/internal/storage"
	"github.com/TheTempest7/social_network/app-go/internal/workers"

	// "github.com/TheTempest7/social_network/app-go/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	// ... твои пакеты auth, handlers, middleware ...
)

func NewRouter(
	db *sql.DB, storage *storage.Storage,
) http.Handler {

	// 👇 Создаём слои
	repo := repository.NewUserRepository(db)

	emailQueue := make(chan workers.EmailTask, 100)
	workers.StartEmailWorkers(3, emailQueue, repo)

	go workers.StartEmailCron(repo, emailQueue)

	userService := user.NewUserService(repo, emailQueue)
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
		json.NewEncoder(w).Encode(map[string]string{"message": "Hello, World!"})
	})

	r.Get("/ping", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "pong"})
	})

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/logout", authHandler.Logout)
	})

	r.Route("/api/users", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return appmw.RequireAuth(blacklist, next) // твой appmw
		})
		r.Get("/profile", profileHandler.Profile)
		r.Put("/profile", profileHandler.PutProfile)
		r.Put("/avatar", profileHandler.UploadAvatar)
		r.Post("/avatar/presign", profileHandler.PresignAvatar)
		r.Post("/avatar/confirm", profileHandler.ConfirmAvatar)
	})

	return r
}
