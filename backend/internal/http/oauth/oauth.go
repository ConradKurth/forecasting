package oauth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ConradKurth/forecasting/backend/internal/auth"
	"github.com/ConradKurth/forecasting/backend/internal/config"
	"github.com/go-chi/chi/v5"
)

func InitRoutes(r *chi.Mux) {
	r.Get("/v1/shopify/install", RequestInstall)
	r.Get("/v1/shopify/callback", RequestCallback)
}

func RequestInstall(w http.ResponseWriter, r *http.Request) {
	shop := r.URL.Query().Get("shop")
	if shop == "" {
		http.Error(w, "Missing shop parameter", http.StatusBadRequest)
		return
	}
	fmt.Println("Redirecting to", config.Values.Shopify.RedirectURL)
	redirectURL := fmt.Sprintf("https://%s/admin/oauth/authorize?client_id=%s&scope=%s&redirect_uri=%s",
		shop, config.Values.Shopify.ClientID, url.QueryEscape(strings.Join(config.Values.Shopify.Scopes, ",")), url.QueryEscape(config.Values.Shopify.RedirectURL))

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func RequestCallback(w http.ResponseWriter, r *http.Request) {
	shop := r.URL.Query().Get("shop")
	code := r.URL.Query().Get("code")
	fmt.Println("Getting the callback")
	if shop == "" || code == "" {
		http.Error(w, "Missing shop or code", http.StatusBadRequest)
		return
	}

	accessTokenURL := fmt.Sprintf("https://%s/admin/oauth/access_token", shop)

	form := url.Values{}
	form.Set("client_id", config.Values.Shopify.ClientID)
	form.Set("client_secret", config.Values.Shopify.ClientSecret)
	form.Set("code", code)

	resp, err := http.Post(accessTokenURL, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		http.Error(w, "Failed to get access token", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		http.Error(w, "Failed to decode access token", http.StatusInternalServerError)
		return
	}

	// Generate JWT token for the authenticated user
	jwtToken, err := auth.GenerateJWT(shop, shop) // Using shop as userID for now
	if err != nil {
		http.Error(w, "Failed to generate JWT token", http.StatusInternalServerError)
		return
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
	fmt.Println("Redirecting to", config.Values.Frontend.URL)
	redirectTo := fmt.Sprintf("%v/callback?shop=%s&token=%s", config.Values.Frontend.URL, shop, jwtToken)
	http.Redirect(w, r, redirectTo, http.StatusFound)
}
