package logger

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type logKey struct{}

// TODO: add logging for websocket

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
}

func SlogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		reqLog := slog.Default().With(
			"method", r.Method,
			"path", r.URL.Path,
		)

		ctx := context.WithValue(r.Context(), logKey{}, reqLog)
		r = r.WithContext(ctx)

		start := time.Now()
		next.ServeHTTP(w, r)
		reqLog.Info("request complete", "ms", time.Since(start).Milliseconds())
	})
}

func GetLog(ctx context.Context) *slog.Logger {
	l, ok := ctx.Value(logKey{}).(*slog.Logger)
	if ok {
		return l
	}

	return slog.Default()
}
