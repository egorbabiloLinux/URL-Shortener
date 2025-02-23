package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"URL-Shortener/internal/lib/logger/sl"

	"github.com/golang-jwt/jwt/v5"
)

type PermissionProvider interface {
	IsAdmin(
		ctx context.Context,
		userID int64,
	) (bool, error)
}

var (
	ErrInvalidToken = fmt.Errorf("invalid token")
	ErrIsAdminCheckFailed = fmt.Errorf("failed to check if user is admin")
)

func New(
	log *slog.Logger,
	appSecret string,
	permProvider PermissionProvider,
) func(next http.Handler) http.Handler {
	const op = "middleware.auth.New"

	log = log.With(slog.String("op", op))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
			tokenStr := ExtractBearerToken(r)
			if tokenStr == "" {
				// It's ok, if user is not authorized
				next.ServeHTTP(w, r)
				return
			}

			tokenParsed, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				return []byte(appSecret), nil
			})
			if err != nil {
				log.Warn("failed to parse token", sl.Err(err))

				// But if token is invalid, we shouldn't handle request
				ctx := context.WithValue(r.Context(), errorKey, ErrInvalidToken)
				next.ServeHTTP(w, r.WithContext(ctx))

				return
			}
			claims, ok := tokenParsed.Claims.(jwt.MapClaims)
			if !ok {
				log.Warn("failed to cast claims") 
			} else {
				log.Info("user authorized", slog.Any("claims", claims))
			}

			uidFloat, ok := claims["uid"].(float64)
			if !ok {
				log.Warn("missing or invalid uid token")
				ctx := context.WithValue(r.Context(), errorKey, ErrInvalidToken)
				next.ServeHTTP(w, r.WithContext(ctx))

				return
			}

			uid := int64(uidFloat)

			isAdmin, err := permProvider.IsAdmin(r.Context(), uid)
			if err != nil {
				log.Error("failed to check if user is admin", sl.Err(err))

				ctx := context.WithValue(r.Context(), errorKey, ErrIsAdminCheckFailed)
				next.ServeHTTP(w, r.WithContext(ctx))

				return 
			}

			ctx := context.WithValue(r.Context(), uidKey, uid)
			ctx = context.WithValue(ctx, isAdminKey, isAdmin)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ExtractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	splitToken := strings.Split(authHeader, "Bearer ")
	if len(splitToken) != 2 {
		return ""
	}

	return splitToken[1]
}

func UIDFromContext(ctx context.Context) (int64, bool) {
	uid, ok := ctx.Value(uidKey).(int64)
	return uid, ok
}

func IsAdminFromContext(ctx context.Context) (bool, bool) {
	isAdmin, ok := ctx.Value(isAdminKey).(bool)
	return isAdmin, ok
}

func ErrorFromContext(ctx context.Context) (error, bool) {
	err, ok := ctx.Value(errorKey).(error)
	return err, ok
}
