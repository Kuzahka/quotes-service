package domain_test

import (
	"testing"

	"quotes-service/internal/domain"
)

func TestCreateQuoteRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     domain.CreateQuoteRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: domain.CreateQuoteRequest{
				Author: "Test Author",
				Quote:  "Test quote text",
			},
			wantErr: false,
		},
		{
			name: "empty author",
			req: domain.CreateQuoteRequest{
				Author: "",
				Quote:  "Test quote text",
			},
			wantErr: true,
			errMsg:  "author is required",
		},
		{
			name: "empty quote",
			req: domain.CreateQuoteRequest{
				Author: "Test Author",
				Quote:  "",
			},
			wantErr: true,
			errMsg:  "quote is required",
		},
		{
			name: "whitespace only author",
			req: domain.CreateQuoteRequest{
				Author: "   ",
				Quote:  "Test quote text",
			},
			wantErr: true,
			errMsg:  "author is required",
		},
		{
			name: "whitespace only quote",
			req: domain.CreateQuoteRequest{
				Author: "Test Author",
				Quote:  "   ",
			},
			wantErr: true,
			errMsg:  "quote is required",
		},
		{
			name: "author too long",
			req: domain.CreateQuoteRequest{
				Author: string(make([]rune, 101)), // 101 characters
				Quote:  "Test quote text",
			},
			wantErr: true,
			errMsg:  "author must be less than 100 characters",
		},
		{
			name: "quote too long",
			req: domain.CreateQuoteRequest{
				Author: "Test Author",
				Quote:  string(make([]rune, 1001)), // 1001 characters
			},
			wantErr: true,
			errMsg:  "quote must be less than 1000 characters",
		},
		{
			name: "trims whitespace",
			req: domain.CreateQuoteRequest{
				Author: "  Test Author  ",
				Quote:  "  Test quote text  ",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				// Check that whitespace was trimmed
				if tt.req.Author != "Test Author" || tt.req.Quote != "Test quote text" {
					t.Errorf("Expected whitespace to be trimmed")
				}
			}
		})
	}
}
