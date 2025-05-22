package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"quotes-service/internal/domain"
	"quotes-service/internal/infrastructure/logger"
	"quotes-service/internal/service"

	"github.com/gorilla/mux"
)

type QuoteHandler struct {
	service *service.QuoteService
	logger  *logger.Logger
}

type Response struct {
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Database  string    `json:"database"`
	Uptime    string    `json:"uptime"`
}

var startTime = time.Now()

func NewQuoteHandler(service *service.QuoteService, logger *logger.Logger) *QuoteHandler {
	return &QuoteHandler{
		service: service,
		logger:  logger,
	}
}

func (h *QuoteHandler) RegisterRoutes(router *mux.Router) {
	// API routes
	router.HandleFunc("/quotes", h.CreateQuote).Methods("POST")
	router.HandleFunc("/quotes", h.GetQuotes).Methods("GET")
	router.HandleFunc("/quotes/random", h.GetRandomQuote).Methods("GET")
	router.HandleFunc("/quotes/{id:[0-9]+}", h.DeleteQuote).Methods("DELETE")

	// Health check
	router.HandleFunc("/health", h.HealthCheck).Methods("GET")

	// Add middleware
	router.Use(h.loggingMiddleware)
	router.Use(h.recoveryMiddleware)
}

func (h *QuoteHandler) CreateQuote(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var req domain.CreateQuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Debug("Invalid JSON in request", "error", err)
		h.sendError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	quote, err := h.service.CreateQuote(ctx, req)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidQuote) {
			h.sendError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error("Failed to create quote", "error", err)
		h.sendError(w, http.StatusInternalServerError, "Failed to create quote")
		return
	}

	h.sendSuccess(w, http.StatusCreated, quote)
}

func (h *QuoteHandler) GetQuotes(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	filter := domain.QuoteFilter{
		Author: r.URL.Query().Get("author"),
	}

	// Parse limit parameter
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	// Parse offset parameter
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	quotes, err := h.service.GetAllQuotes(ctx, filter)
	if err != nil {
		h.logger.Error("Failed to get quotes", "error", err, "filter", filter)
		h.sendError(w, http.StatusInternalServerError, "Failed to get quotes")
		return
	}

	h.sendSuccess(w, http.StatusOK, quotes)
}

func (h *QuoteHandler) GetRandomQuote(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	quote, err := h.service.GetRandomQuote(ctx)
	if err != nil {
		if errors.Is(err, domain.ErrQuoteNotFound) {
			h.sendError(w, http.StatusNotFound, "No quotes found")
			return
		}
		h.logger.Error("Failed to get random quote", "error", err)
		h.sendError(w, http.StatusInternalServerError, "Failed to get random quote")
		return
	}

	h.sendSuccess(w, http.StatusOK, quote)
}

func (h *QuoteHandler) DeleteQuote(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid quote ID")
		return
	}

	err = h.service.DeleteQuote(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrQuoteNotFound) {
			h.sendError(w, http.StatusNotFound, "Quote not found")
			return
		}
		if errors.Is(err, domain.ErrInvalidQuote) {
			h.sendError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error("Failed to delete quote", "id", id, "error", err)
		h.sendError(w, http.StatusInternalServerError, "Failed to delete quote")
		return
	}

	h.sendSuccess(w, http.StatusOK, map[string]string{
		"message": "Quote deleted successfully",
	})
}

func (h *QuoteHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	dbStatus := "connected"
	if err := h.service.HealthCheck(ctx); err != nil {
		h.logger.Error("Database health check failed", "error", err)
		dbStatus = "disconnected"
	}

	uptime := time.Since(startTime).String()

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Database:  dbStatus,
		Uptime:    uptime,
	}

	// If database is down, return 503
	if dbStatus == "disconnected" {
		response.Status = "unhealthy"
		h.sendResponse(w, http.StatusServiceUnavailable, Response{Data: response})
		return
	}

	h.sendSuccess(w, http.StatusOK, response)
}

func (h *QuoteHandler) sendSuccess(w http.ResponseWriter, statusCode int, data interface{}) {
	h.sendResponse(w, statusCode, Response{Data: data})
}

func (h *QuoteHandler) sendError(w http.ResponseWriter, statusCode int, message string) {
	h.sendResponse(w, statusCode, Response{Error: message})
}

func (h *QuoteHandler) sendResponse(w http.ResponseWriter, statusCode int, response Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

// Middleware for logging HTTP requests
func (h *QuoteHandler) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap ResponseWriter to capture status code
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(ww, r)

		duration := time.Since(start)

		h.logger.Info("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.statusCode,
			"duration", duration.String(),
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)
	})
}

// Middleware for panic recovery
func (h *QuoteHandler) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				h.logger.Error("Panic recovered",
					"error", err,
					"path", r.URL.Path,
					"method", r.Method,
				)

				h.sendError(w, http.StatusInternalServerError, "Internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
