package main

import (
	"URL-Shortener/internal/client/sso/grpc"
	"URL-Shortener/internal/config" // Путь к пакету config
	"URL-Shortener/internal/http-server/handlers/auth/register"
	del "URL-Shortener/internal/http-server/handlers/url/delete"
	"URL-Shortener/internal/http-server/handlers/url/redirect"
	"URL-Shortener/internal/http-server/handlers/url/save"
	"URL-Shortener/internal/http-server/middleware/auth"
	mwLogger "URL-Shortener/internal/http-server/middleware/logger"
	"URL-Shortener/internal/lib/logger/sl"
	"URL-Shortener/internal/storage/postgres"
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)
	log = log.With(slog.String("env", cfg.Env))

	log.Info("Initializing server", slog.String("address", cfg.Address))
	log.Debug("logger debug mode enabled")

	storage, err := postgres.NewStorage("postgres://postgres:3547@localhost:5432/urlshortener")
	if err != nil {
		log.Error("failed to initialize storage", sl.Err(err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	grpcAuth, err := grpc.New(ctx, log, cfg.SSOGrpcAddr, cfg.SSOGrpcTimeout, 3)
	if err != nil {
		log.Error("failed to initialize gRPC permission provider", sl.Err(err))

		os.Exit(1)
	}

	authMiddleware := auth.New(log, cfg.AppSecret, permProvider)
	
	router := chi.NewRouter()
  
	router.Use(middleware.RequestID) /* Добавляет request_id в каждый запрос, для трейсинга, 
	чтобы отслеживать запрос через разные части системы (например, логирование, базы данных, микросервисы). */

	router.Use(mwLogger.New(log)) // Логирование всех запросов

	router.Use(middleware.Recoverer) /* Если где-то внутри сервера (обработчика запроса) произойдет паника, приложение не должно упасть. 
	Перехватывает панику внутри обработчиков запросов и возвращает HTTP-ответ с кодом ошибки (обычно 500 Internal Server Error). */

	router.Use(middleware.URLFormat) /* Парсер URLов поступающих запросов. 
	Например, если запрос содержит расширения (/api/resource.json), это будет обработано корректно (удалит .json, .xml и передаст "чистый" URL).
	Полезно для обработки запросов, где формат (например, JSON, XML) определяет, как должен выглядеть ответ. */

	router.Route("/url", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Post("/", save.New(log, storage)) // для POST /url
		r.Delete("/{alias}", del.New(log, storage)) // для DELETE /url/{alias}
	})

	router.Get("/{alias}", redirect.New(log, storage))

	router.Post("/register", register.New(log, grpcAuth))

	server := http.Server {
		Addr: cfg.Address,
		Handler: router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.IdleTimeout,
	}

	log.Info("Starting server", slog.String("port", cfg.Address))
	if err := server.ListenAndServe(); err != nil {
		log.Error("failed to start server", sl.Err(err))

		return
	}
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
