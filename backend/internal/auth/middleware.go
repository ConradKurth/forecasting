package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/ConradKurth/forecasting/backend/internal/http/response"
)

type contextKey string

const UserContextKey contextKey = "user"

type User struct {
	Shop   string
	UserID string
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			sendError(w, response.Unauthorized("Authorization header required", nil))
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			sendError(w, response.Unauthorized("Invalid authorization header format", nil))
			return
		}

		claims, err := ValidateJWT(bearerToken[1])
		if err != nil {
			sendError(w, response.InvalidToken())
			return
		}

		user := &User{
			Shop:   claims.Shop,
			UserID: claims.UserID,
		}

		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// sendError sends an HTTP error using the response package utilities
func sendError(w http.ResponseWriter, httpErr *response.HTTPError) {
	errorResp := response.ErrorResponse{
		Error:   http.StatusText(httpErr.StatusCode),
		Message: httpErr.Message,
		Code:    httpErr.ErrorCode,
	}

	// Use the response package's JSON helper which handles headers and status code
	response.JSON(w, httpErr.StatusCode, errorResp)
}

func GetUserFromContext(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(UserContextKey).(*User)
	return user, ok
}
