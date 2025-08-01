package oauth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/ConradKurth/forecasting/backend/internal/auth"
	"github.com/ConradKurth/forecasting/backend/internal/config"
	"github.com/ConradKurth/forecasting/backend/internal/http/response"
	"github.com/ConradKurth/forecasting/backend/internal/service"
	"github.com/go-chi/chi/v5"
)

func InitRoutes(r *chi.Mux, userService *service.UserService) {
	r.Get("/v1/shopify/install", response.Wrap(RequestInstall))
	r.Get("/v1/shopify/callback", response.Wrap(RequestCallback(userService)))
}

func RequestInstall(w http.ResponseWriter, r *http.Request) error {
	shop := r.URL.Query().Get("shop")
	if shop == "" {
		return response.MissingParameter("shop")
	}
	redirectURL := fmt.Sprintf("https://%s/admin/oauth/authorize?client_id=%s&scope=%s&redirect_uri=%s",
		shop, config.Values.Shopify.ClientID, url.QueryEscape(strings.Join(config.Values.Shopify.Scopes, ",")), url.QueryEscape(config.Values.Shopify.RedirectURL))

	http.Redirect(w, r, redirectURL, http.StatusFound)
	return nil
}

func RequestCallback(userService *service.UserService) response.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		shop := r.URL.Query().Get("shop")
		code := r.URL.Query().Get("code")

		if shop == "" {
			return response.MissingParameter("shop")
		}
		if code == "" {
			return response.MissingParameter("code")
		}

		accessTokenURL := fmt.Sprintf("https://%s/admin/oauth/access_token", shop)

		form := url.Values{}
		form.Set("client_id", config.Values.Shopify.ClientID)
		form.Set("client_secret", config.Values.Shopify.ClientSecret)
		form.Set("code", code)

		resp, err := http.Post(accessTokenURL, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
		if err != nil {
			return response.InternalServerError("Failed to get access token", err)
		}
		defer resp.Body.Close()

		var tokenResp struct {
			AccessToken string `json:"access_token"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
			return response.InternalServerError("Failed to decode access token", err)
		}

		// Create or update user with the access token
		user, err := userService.CreateOrUpdateUser(r.Context(), shop, tokenResp.AccessToken)
		if err != nil {
			log.Printf("Failed to create or update user for shop %s: %v", shop, err)
			return response.DatabaseError(err)
		}

		// Generate JWT token for the authenticated user
		jwtToken, err := auth.GenerateJWT(shop, user.ID)
		if err != nil {
			return response.InternalServerError("Failed to generate JWT token", err)
		}

		// Set JWT as an HTTP-only cookie and also return it as JSON
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    jwtToken,
			HttpOnly: true,
			Secure:   false, // Set to true in production with HTTPS
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
			MaxAge:   86400, // 24 hours
		})

		// Redirect to frontend callback page with parameters
		redirectTo := fmt.Sprintf("%v/callback?shop=%s&token=%s", config.Values.Frontend.URL, shop, jwtToken)
		http.Redirect(w, r, redirectTo, http.StatusFound)
		return nil
	}
}
