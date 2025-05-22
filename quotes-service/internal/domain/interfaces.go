package domain

import "context"

type QuoteService interface {
	CreateQuote(ctx context.Context, req CreateQuoteRequest) (*Quote, error)
	GetAllQuotes(ctx context.Context, filter QuoteFilter) ([]*Quote, error)
	GetRandomQuote(ctx context.Context) (*Quote, error)
	DeleteQuote(ctx context.Context, id int) error
	HealthCheck(ctx context.Context) error
}
