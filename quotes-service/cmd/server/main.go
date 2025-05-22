package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"quotes-service/internal/config"
	"quotes-service/internal/handler"
	"quotes-service/internal/infrastructure/database"
	"quotes-service/internal/infrastructure/logger"
	"quotes-service/internal/repository/postgres"
	"quotes-service/internal/service"

	"github.com/gorilla/mux"
)

func main() {
	// Загрузка конфигурации
	cfg := config.Load()

	// Иниициализация логгера
	logger := logger.New(cfg.LogLevel)
	logger.Info("Starting quotes service", "version", "1.0.0")

	// Инициализация базы данных с connection pool
	db, err := database.NewPostgresConnection(cfg.DatabaseConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Error closing database connection", "error", err)
		}
	}()

	logger.Info("Database connection established")

	// Инициализация репозитория
	quoteRepo := postgres.NewQuoteRepository(db, logger)

	// Инициализация сервиса
	quoteService := service.NewQuoteService(quoteRepo, logger)

	// Инициализация хендлера
	quoteHandler := handler.NewQuoteHandler(quoteService, logger)

	// Настройки маршрутизатора
	router := mux.NewRouter()
	quoteHandler.RegisterRoutes(router)

	// Настройка сервера с тайм-аутами
	server := &http.Server{
		Addr:         cfg.ServerAddress,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запуск сервера в отдельной горутине
	serverError := make(chan error, 1)
	go func() {
		logger.Info("HTTP server starting", "address", cfg.ServerAddress)
		serverError <- server.ListenAndServe()
	}()

	// Ожидание сигнала прерывания или ошибки сервера
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverError:
		logger.Error("Server error", "error", err)
	case sig := <-quit:
		logger.Info("Received shutdown signal", "signal", sig.String())
	}

	// Graceful shutdown
	logger.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown error", "error", err)
		os.Exit(1)
	}

	logger.Info("Server shutdown completed")
}
