package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"quotes-service/internal/domain"
	"quotes-service/internal/infrastructure/logger"
	"quotes-service/internal/service"
)

// Mock repository for testing
type mockQuoteRepository struct {
	quotes  []*domain.Quote
	nextID  int
	errOnOp map[string]error
}

func newMockQuoteRepository() *mockQuoteRepository {
	return &mockQuoteRepository{
		quotes:  make([]*domain.Quote, 0),
		nextID:  1,
		errOnOp: make(map[string]error),
	}
}

func (m *mockQuoteRepository) Create(ctx context.Context, quote *domain.Quote) (*domain.Quote, error) {
	if err := m.errOnOp["create"]; err != nil {
		return nil, err
	}

	now := time.Now()
	newQuote := &domain.Quote{
		ID:        m.nextID,
		Author:    quote.Author,
		Text:      quote.Text,
		CreatedAt: now,
		UpdatedAt: now,
	}
	m.nextID++
	m.quotes = append(m.quotes, newQuote)
	return newQuote, nil
}

func (m *mockQuoteRepository) GetAll(ctx context.Context, filter domain.QuoteFilter) ([]*domain.Quote, error) {
	if err := m.errOnOp["getall"]; err != nil {
		return nil, err
	}

	result := make([]*domain.Quote, 0)
	for _, quote := range m.quotes {
		if filter.Author != "" && quote.Author != filter.Author {
			continue
		}
		result = append(result, quote)
	}

	// Apply limit and offset
	if filter.Offset > 0 && filter.Offset < len(result) {
		result = result[filter.Offset:]
	} else if filter.Offset >= len(result) {
		result = []*domain.Quote{}
	}

	if filter.Limit > 0 && filter.Limit < len(result) {
		result = result[:filter.Limit]
	}

	return result, nil
}

func (m *mockQuoteRepository) GetByID(ctx context.Context, id int) (*domain.Quote, error) {
	if err := m.errOnOp["getbyid"]; err != nil {
		return nil, err
	}

	for _, quote := range m.quotes {
		if quote.ID == id {
			return quote, nil
		}
	}
	return nil, domain.ErrQuoteNotFound
}

func (m *mockQuoteRepository) GetRandom(ctx context.Context) (*domain.Quote, error) {
	if err := m.errOnOp["getrandom"]; err != nil {
		return nil, err
	}

	if len(m.quotes) == 0 {
		return nil, domain.ErrQuoteNotFound
	}
	return m.quotes[0], nil // Just return first for simplicity
}

func (m *mockQuoteRepository) Delete(ctx context.Context, id int) error {
	if err := m.errOnOp["delete"]; err != nil {
		return err
	}

	for i, quote := range m.quotes {
		if quote.ID == id {
			m.quotes = append(m.quotes[:i], m.quotes[i+1:]...)
			return nil
		}
	}
	return domain.ErrQuoteNotFound
}

func (m *mockQuoteRepository) Count(ctx context.Context, filter domain.QuoteFilter) (int, error) {
	if err := m.errOnOp["count"]; err != nil {
		return 0, err
	}

	count := 0
	for _, quote := range m.quotes {
		if filter.Author != "" && quote.Author != filter.Author {
			continue
		}
		count++
	}
	return count, nil
}

func (m *mockQuoteRepository) HealthCheck(ctx context.Context) error {
	return m.errOnOp["healthcheck"]
}

func TestQuoteService_CreateQuote(t *testing.T) {
	mockRepo := newMockQuoteRepository()
	logger := logger.New("debug")
	service := service.NewQuoteService(mockRepo, logger)

	tests := []struct {
		name    string
		req     domain.CreateQuoteRequest
		wantErr bool
		setup   func()
	}{
		{
			name: "valid quote creation",
			req: domain.CreateQuoteRequest{
				Author: "Test Author",
				Quote:  "Test quote",
			},
			wantErr: false,
		},
		{
			name: "invalid quote - empty author",
			req: domain.CreateQuoteRequest{
				Author: "",
				Quote:  "Test quote",
			},
			wantErr: true,
		},
		{
			name: "repository error",
			req: domain.CreateQuoteRequest{
				Author: "Test Author",
				Quote:  "Test quote",
			},
			wantErr: true,
			setup: func() {
				mockRepo.errOnOp["create"] = errors.New("database error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			ctx := context.Background()
			quote, err := service.CreateQuote(ctx, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if quote == nil {
					t.Errorf("Expected quote but got nil")
				}
				if quote != nil && quote.Author != tt.req.Author {
					t.Errorf("Expected author '%s', got '%s'", tt.req.Author, quote.Author)
				}
			}

			// Reset mock for next test
			mockRepo.errOnOp = make(map[string]error)
		})
	}
}

func TestQuoteService_GetAllQuotes(t *testing.T) {
	mockRepo := newMockQuoteRepository()
	logger := logger.New("debug")
	service := service.NewQuoteService(mockRepo, logger)

	// Add some test data
	testQuotes := []*domain.Quote{
		{ID: 1, Author: "Author 1", Text: "Quote 1"},
		{ID: 2, Author: "Author 2", Text: "Quote 2"},
		{ID: 3, Author: "Author 1", Text: "Quote 3"},
	}
	mockRepo.quotes = testQuotes
	mockRepo.nextID = 4

	tests := []struct {
		name      string
		filter    domain.QuoteFilter
		wantCount int
		wantErr   bool
		setup     func()
	}{
		{
			name:      "get all quotes",
			filter:    domain.QuoteFilter{},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "filter by author",
			filter:    domain.QuoteFilter{Author: "Author 1"},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "limit results",
			filter:    domain.QuoteFilter{Limit: 2},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "default limit applied",
			filter:    domain.QuoteFilter{Limit: 0}, // Should get default limit of 100
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:    "repository error",
			filter:  domain.QuoteFilter{},
			wantErr: true,
			setup: func() {
				mockRepo.errOnOp["getall"] = errors.New("database error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			ctx := context.Background()
			quotes, err := service.GetAllQuotes(ctx, tt.filter)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if len(quotes) != tt.wantCount {
					t.Errorf("Expected %d quotes, got %d", tt.wantCount, len(quotes))
				}
			}

			// Reset mock for next test
			mockRepo.errOnOp = make(map[string]error)
		})
	}
}

func TestQuoteService_GetRandomQuote(t *testing.T) {
	mockRepo := newMockQuoteRepository()
	logger := logger.New("debug")
	service := service.NewQuoteService(mockRepo, logger)

	tests := []struct {
		name    string
		setup   func()
		wantErr bool
	}{
		{
			name: "get random quote success",
			setup: func() {
				mockRepo.quotes = []*domain.Quote{
					{ID: 1, Author: "Test Author", Text: "Test Quote"},
				}
			},
			wantErr: false,
		},
		{
			name: "no quotes available",
			setup: func() {
				mockRepo.quotes = []*domain.Quote{}
				mockRepo.errOnOp["getrandom"] = domain.ErrQuoteNotFound
			},
			wantErr: true,
		},
		{
			name: "repository error",
			setup: func() {
				mockRepo.errOnOp["getrandom"] = errors.New("database error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			ctx := context.Background()
			quote, err := service.GetRandomQuote(ctx)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if quote == nil {
					t.Errorf("Expected quote but got nil")
				}
			}

			// Reset mock for next test
			mockRepo.errOnOp = make(map[string]error)
			mockRepo.quotes = []*domain.Quote{}
		})
	}
}

func TestQuoteService_DeleteQuote(t *testing.T) {
	mockRepo := newMockQuoteRepository()
	logger := logger.New("debug")
	service := service.NewQuoteService(mockRepo, logger)

	tests := []struct {
		name    string
		id      int
		setup   func()
		wantErr bool
	}{
		{
			name: "successful deletion",
			id:   1,
			setup: func() {
				mockRepo.quotes = []*domain.Quote{
					{ID: 1, Author: "Test Author", Text: "Test Quote"},
				}
			},
			wantErr: false,
		},
		{
			name:    "invalid ID",
			id:      0,
			wantErr: true,
		},
		{
			name:    "negative ID",
			id:      -1,
			wantErr: true,
		},
		{
			name: "quote not found",
			id:   999,
			setup: func() {
				mockRepo.errOnOp["delete"] = domain.ErrQuoteNotFound
			},
			wantErr: true,
		},
		{
			name: "repository error",
			id:   1,
			setup: func() {
				mockRepo.errOnOp["delete"] = errors.New("database error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			ctx := context.Background()
			err := service.DeleteQuote(ctx, tt.id)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}

			// Reset mock for next test
			mockRepo.errOnOp = make(map[string]error)
			mockRepo.quotes = []*domain.Quote{}
		})
	}
}

func TestQuoteService_HealthCheck(t *testing.T) {
	mockRepo := newMockQuoteRepository()
	logger := logger.New("debug")
	service := service.NewQuoteService(mockRepo, logger)

	tests := []struct {
		name    string
		setup   func()
		wantErr bool
	}{
		{
			name:    "health check success",
			wantErr: false,
		},
		{
			name: "health check failure",
			setup: func() {
				mockRepo.errOnOp["healthcheck"] = errors.New("database connection failed")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			ctx := context.Background()
			err := service.HealthCheck(ctx)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}

			// Reset mock for next test
			mockRepo.errOnOp = make(map[string]error)
		})
	}
}
