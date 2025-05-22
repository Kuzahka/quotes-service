package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"quotes-service/internal/domain"
	"quotes-service/internal/infrastructure/logger"
)

type quoteRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

func NewQuoteRepository(db *sql.DB, logger *logger.Logger) domain.QuoteRepository {
	return &quoteRepository{
		db:     db,
		logger: logger,
	}
}

func (r *quoteRepository) Create(ctx context.Context, quote *domain.Quote) (*domain.Quote, error) {
	query := `
		INSERT INTO quotes (author, text, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, author, text, created_at, updated_at`

	now := time.Now()
	quote.CreatedAt = now
	quote.UpdatedAt = now

	var result domain.Quote
	err := r.db.QueryRowContext(ctx, query, quote.Author, quote.Text, now, now).Scan(
		&result.ID, &result.Author, &result.Text, &result.CreatedAt, &result.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create quote", "error", err, "author", quote.Author)
		return nil, fmt.Errorf("failed to create quote: %w", err)
	}

	r.logger.Info("Quote created", "id", result.ID, "author", result.Author)
	return &result, nil
}

func (r *quoteRepository) GetAll(ctx context.Context, filter domain.QuoteFilter) ([]*domain.Quote, error) {
	query := "SELECT id, author, text, created_at, updated_at FROM quotes"
	args := []interface{}{}
	conditions := []string{}

	if filter.Author != "" {
		conditions = append(conditions, "author ILIKE $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, "%"+filter.Author+"%")
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", len(args)+1)
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", len(args)+1)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to get quotes", "error", err, "filter", filter)
		return nil, fmt.Errorf("failed to get quotes: %w", err)
	}
	defer rows.Close()

	var quotes []*domain.Quote
	for rows.Next() {
		var quote domain.Quote
		err := rows.Scan(&quote.ID, &quote.Author, &quote.Text, &quote.CreatedAt, &quote.UpdatedAt)
		if err != nil {
			r.logger.Error("Failed to scan quote", "error", err)
			return nil, fmt.Errorf("failed to scan quote: %w", err)
		}
		quotes = append(quotes, &quote)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over quotes: %w", err)
	}

	r.logger.Debug("Retrieved quotes", "count", len(quotes), "filter", filter)
	return quotes, nil
}

func (r *quoteRepository) GetByID(ctx context.Context, id int) (*domain.Quote, error) {
	query := "SELECT id, author, text, created_at, updated_at FROM quotes WHERE id = $1"

	var quote domain.Quote
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&quote.ID, &quote.Author, &quote.Text, &quote.CreatedAt, &quote.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrQuoteNotFound
		}
		r.logger.Error("Failed to get quote by ID", "error", err, "id", id)
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}

	return &quote, nil
}

func (r *quoteRepository) GetRandom(ctx context.Context) (*domain.Quote, error) {
	query := "SELECT id, author, text, created_at, updated_at FROM quotes ORDER BY RANDOM() LIMIT 1"

	var quote domain.Quote
	err := r.db.QueryRowContext(ctx, query).Scan(
		&quote.ID, &quote.Author, &quote.Text, &quote.CreatedAt, &quote.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrQuoteNotFound
		}
		r.logger.Error("Failed to get random quote", "error", err)
		return nil, fmt.Errorf("failed to get random quote: %w", err)
	}

	r.logger.Debug("Retrieved random quote", "id", quote.ID, "author", quote.Author)
	return &quote, nil
}

func (r *quoteRepository) Delete(ctx context.Context, id int) error {
	query := "DELETE FROM quotes WHERE id = $1"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete quote", "error", err, "id", id)
		return fmt.Errorf("failed to delete quote: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrQuoteNotFound
	}

	r.logger.Info("Quote deleted", "id", id)
	return nil
}

func (r *quoteRepository) Count(ctx context.Context, filter domain.QuoteFilter) (int, error) {
	query := "SELECT COUNT(*) FROM quotes"
	args := []interface{}{}
	conditions := []string{}

	if filter.Author != "" {
		conditions = append(conditions, "author ILIKE $1")
		args = append(args, "%"+filter.Author+"%")
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		r.logger.Error("Failed to count quotes", "error", err, "filter", filter)
		return 0, fmt.Errorf("failed to count quotes: %w", err)
	}

	return count, nil
}

func (r *quoteRepository) HealthCheck(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
