package domain

import "context"

type QuoteRepository interface {
	Create(ctx context.Context, quote *Quote) (*Quote, error)
	GetAll(ctx context.Context, filter QuoteFilter) ([]*Quote, error)
	GetByID(ctx context.Context, id int) (*Quote, error)
	GetRandom(ctx context.Context) (*Quote, error)
	Delete(ctx context.Context, id int) error
	Count(ctx context.Context, filter QuoteFilter) (int, error)
	HealthCheck(ctx context.Context) error
}
