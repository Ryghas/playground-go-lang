package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pokemon-service/internal/delivery/http/handler"

	"github.com/gin-gonic/gin"
)

type fakeChecker struct{ err error }

func (f fakeChecker) Ping(context.Context) error { return f.err }

func TestHealth(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		pingErr  error
		status   int
		contains string
	}{
		{"live selalu ok walau db mati", "/health/live", errors.New("db down"), http.StatusOK, `"status":"ok"`},
		{"ready db up", "/health/ready", nil, http.StatusOK, `"database":"up"`},
		{"ready db down", "/health/ready", errors.New("db down"), http.StatusServiceUnavailable, `"database":"down"`},
		{"alias /health = ready", "/health", errors.New("db down"), http.StatusServiceUnavailable, `"status":"unavailable"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			handler.NewHealthHandler(fakeChecker{err: tt.pingErr}).Register(r)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if w.Code != tt.status {
				t.Fatalf("status = %d, want %d", w.Code, tt.status)
			}
			if !strings.Contains(w.Body.String(), tt.contains) {
				t.Errorf("body %s tidak mengandung %s", w.Body.String(), tt.contains)
			}
		})
	}
}
