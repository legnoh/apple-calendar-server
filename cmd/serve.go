package cmd

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/legnoh/apple-calendar-server/pkg/applecalendar"
)

// ServeCmd holds configuration for the server subcommand.
type ServeCmd struct {
	Listen    string   `help:"HTTP listen address" default:":8080" yaml:"listen"`
	Calendars []string `help:"List of calendar names" yaml:"calendars"`
}

func (c *ServeCmd) Run() error { return runServer(c) }

// loggingMiddleware wraps an http.Handler and logs access information
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Log the access information
		duration := time.Since(start)
		slog.Info("access",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("query", r.URL.RawQuery),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
			slog.Int("status", rw.statusCode),
			slog.Duration("duration", duration),
		)
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

func runServer(cfg *ServeCmd) error {
	// Setup structured logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, origin, accept")

		// Extract allowed query parameters
		allowedParams := []string{"from", "to", "limit", "exclude-all-day", "exclude-long-event"}
		params := make(map[string]string)
		query := r.URL.Query()

		for _, param := range allowedParams {
			if values := query[param]; len(values) > 0 {
				params[param] = values[0]
			}
		}

		client := applecalendar.New()
		defer client.Close()
		out, err := client.GetEventList(cfg.Calendars, params)

		if err != nil {
			var ee *exec.ExitError
			w.WriteHeader(http.StatusInternalServerError)
			if errors.As(err, &ee) {
				_, _ = w.Write(ee.Stderr)
				return
			}
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		trimmed := strings.TrimSpace(string(out))
		if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
		} else {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		}

		_, _ = w.Write(out)
	})

	// Apply logging middleware
	handler := loggingMiddleware(mux)

	slog.Info("server starting",
		slog.String("listen", cfg.Listen),
		slog.Any("calendars", cfg.Calendars),
	)

	return http.ListenAndServe(cfg.Listen, handler)
}
