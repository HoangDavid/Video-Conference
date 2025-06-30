package security

import (
	"context"
	"net/http"
	"strings"
)

type ctxKey struct{}

func ClaimsFrom(ctx context.Context) *Claims {
	c, _ := ctx.Value(ctxKey{}).(*Claims)
	return c
}

func IssuerFrom(ctx context.Context) *Issuer {
	i, _ := ctx.Value(ctxKey{}).(*Issuer)
	return i
}

func RequireAuth(i *Issuer) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := extractToken(r)
			if raw == "" {
				http.Error(w, "missing token", http.StatusUnauthorized)
				return
			}

			claims, err := i.Parse(raw)
			if err != nil {
				http.Error(w, "unauthorized: "+err.Error(), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ctxKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))

		})
	}
}

func WithIssuer(i *Issuer) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), ctxKey{}, i)
			next(w, r.WithContext(ctx))
		}
	}
}

func extractToken(r *http.Request) string {
	h := r.Header.Get("Authorization")

	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}

	if q := r.URL.Query().Get("token"); q != "" {
		return q
	}

	return ""

}
