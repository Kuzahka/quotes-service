package service

import (
	"context"
	"fmt"
	"time"

	"quotes-service/internal/domain"
	"quotes-service/internal/infrastructure/logger"
)

type QuoteService struct {
	repo   domain.QuoteRepository
	logger *logger.Logger
}

func NewQuoteService(repo domain.QuoteRepository, logger *logger.Logger) *QuoteService {
	return &QuoteService{
		repo:   repo,
		logger: logger,
	}
}

func (s *QuoteService) CreateQuote(ctx context.Context, req domain.CreateQuoteRequest) (*domain.Quote, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		s.logger.Debug("Invalid quote request", "error", err, "request", req)
		return nil, fmt.Errorf("%w: %s", domain.ErrInvalidQuote, err.Error())
	}

	quote := &domain.Quote{
		Author: req.Author,
		Text:   req.Quote,
	}

	// Добавление метаданных
	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	createdQuote, err := s.repo.Create(dbCtx, quote)
	if err != nil {
		s.logger.Error("Failed to create quote", "error", err, "author", req.Author)
		return nil, fmt.Errorf("failed to create quote: %w", err)
	}

	s.logger.Info("Quote created successfully", "id", createdQuote.ID, "author", createdQuote.Author)
	return createdQuote, nil
}

func (s *QuoteService) GetAllQuotes(ctx context.Context, filter domain.QuoteFilter) ([]*domain.Quote, error) {
	// Установка значений по умолчанию для фильтра
	if filter.Limit <= 0 {
		filter.Limit = 100 // Default limit
	}
	if filter.Limit > 1000 {
		filter.Limit = 1000 // Max limit
	}

	dbCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	quotes, err := s.repo.GetAll(dbCtx, filter)
	if err != nil {
		s.logger.Error("Failed to get quotes", "error", err, "filter", filter)
		return nil, fmt.Errorf("failed to get quotes: %w", err)
	}

	s.logger.Debug("Retrieved quotes", "count", len(quotes), "filter", filter)
	return quotes, nil
}

func (s *QuoteService) GetRandomQuote(ctx context.Context) (*domain.Quote, error) {
	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	quote, err := s.repo.GetRandom(dbCtx)
	if err != nil {
		s.logger.Error("Failed to get random quote", "error", err)
		return nil, fmt.Errorf("failed to get random quote: %w", err)
	}

	s.logger.Debug("Retrieved random quote", "id", quote.ID, "author", quote.Author)
	return quote, nil
}

func (s *QuoteService) DeleteQuote(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("%w: invalid quote ID", domain.ErrInvalidQuote)
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := s.repo.Delete(dbCtx, id)
	if err != nil {
		s.logger.Error("Failed to delete quote", "id", id, "error", err)
		return fmt.Errorf("failed to delete quote: %w", err)
	}

	s.logger.Info("Quote deleted successfully", "id", id)
	return nil
}

func (s *QuoteService) HealthCheck(ctx context.Context) error {
	dbCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return s.repo.HealthCheck(dbCtx)
}
