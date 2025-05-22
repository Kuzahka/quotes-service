package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrQuoteNotFound = errors.New("quote not found")
	ErrInvalidQuote  = errors.New("invalid quote data")
)

type Quote struct {
	ID        int       `json:"id" db:"id"`
	Author    string    `json:"author" db:"author"`
	Text      string    `json:"quote" db:"text"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CreateQuoteRequest struct {
	Author string `json:"author"`
	Quote  string `json:"quote"`
}

func (r *CreateQuoteRequest) Validate() error {
	r.Author = strings.TrimSpace(r.Author)
	r.Quote = strings.TrimSpace(r.Quote)

	if r.Author == "" {
		return errors.New("author is required")
	}
	if r.Quote == "" {
		return errors.New("quote is required")
	}
	if len(r.Author) > 100 {
		return errors.New("author must be less than 100 characters")
	}
	if len(r.Quote) > 1000 {
		return errors.New("quote must be less than 1000 characters")
	}

	return nil
}

type QuoteFilter struct {
	Author string
	Limit  int
	Offset int
}
