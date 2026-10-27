package ollama

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestChat(t *testing.T) {
	tests := []struct {
		name        string
		handler     http.HandlerFunc
		ctxTimeout  time.Duration
		wantContent string
		wantStatus  int    // non-zero means we expect an *APIError with this status
		wantBody    string // text the *APIError body should contain
		wantTimeout bool   // expect context.DeadlineExceeded
		wantErr     bool
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"model":"test","message":{"role":"assistant","content":"hi"}}`)
			},
			ctxTimeout:  time.Second,
			wantContent: "hi",
		},
		{
			name: "server error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, `{"error":"boom"}`)
			},
			ctxTimeout: time.Second,
			wantStatus: http.StatusInternalServerError,
			wantBody:   "boom",
			wantErr:    true,
		},
		{
			name: "malformed json",
			handler: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"message":`)
			},
			ctxTimeout: time.Second,
			wantErr:    true,
		},
		{
			name: "slow server",
			handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(200 * time.Millisecond)
				fmt.Fprint(w, `{"message":{"role":"assistant","content":"too late"}}`)
			},
			ctxTimeout:  50 * time.Millisecond,
			wantTimeout: true,
			wantErr:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(tc.handler)
			defer srv.Close()

			client := CreateClient(srv.URL, 5*time.Second)
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()

			got, err := client.Chat(ctx, ChatRequest{
				Model:    "test",
				Messages: []Message{{Role: "user", Content: "hello"}},
			})

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var apiErr *APIError
			isAPIErr := errors.As(err, &apiErr)
			if tc.wantStatus != 0 {
				if !isAPIErr {
					t.Fatalf("expected *APIError, got %T: %v", err, err)
				}
				if apiErr.StatusCode != tc.wantStatus {
					t.Errorf("status = %d, want %d", apiErr.StatusCode, tc.wantStatus)
				}
				if !strings.Contains(apiErr.Body, tc.wantBody) {
					t.Errorf("body = %q, want it to contain %q", apiErr.Body, tc.wantBody)
				}
			} else if isAPIErr {
				t.Errorf("did not expect *APIError, got %v", apiErr)
			}

			if tc.wantTimeout && !errors.Is(err, context.DeadlineExceeded) {
				t.Errorf("expected context.DeadlineExceeded, got %v", err)
			}

			if got.Message.Content != tc.wantContent {
				t.Errorf("content = %q, want %q", got.Message.Content, tc.wantContent)
			}
		})
	}
}
