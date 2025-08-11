package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"
)

// ====== FACEBOOK (jäta alles, kui kasutad) ======
func facebookOauthConf() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     "222679544825854",
		ClientSecret: "c5560b28c59b7d5afdec1b4d07e51a02",
		RedirectURL:  config.FacebookRedirect,
		Scopes:       []string{"email"},
		Endpoint:     facebook.Endpoint,
	}
}

// ====== GOOGLE ======
func googleOauthConf() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("SUPERWORK_GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("SUPERWORK_GOOGLE_CLIENT_SECRET"),
		RedirectURL:  config.GoogleRedirect, // nt https://superwork-o026.onrender.com/api/oauth2callback/google
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
			"openid",
		},
		Endpoint: google.Endpoint,
	}
}

// /api/google_login_url → tagastab Google'i OAuth-i URL-i
func handleGetGoogleLoginURL(w http.ResponseWriter, r *http.Request) {
	conf := googleOauthConf()
	if conf.ClientID == "" || conf.ClientSecret == "" || conf.RedirectURL == "" {
		http.Error(w, "Google OAuth not configured", http.StatusInternalServerError)
		return
	}
	// Lihtne state (soovi korral salvesta sessiooni ja kontrolli callbackis)
	state := "state"
	url := conf.AuthCodeURL(state, oauth2.AccessTypeOnline)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(url))
}

// /api/oauth2callback/google → vahetab code tokeni vastu ja toob userinfo
func handleGetOauthCallbackGoogle(w http.ResponseWriter, r *http.Request) {
	conf := googleOauthConf()
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		http.Error(w, fmt.Sprintf("token exchange failed: %v", err), http.StatusBadRequest)
		return
	}

	// Küsi kasutaja andmed (OpenID userinfo)
	req, _ := http.NewRequest("GET", "https://openidconnect.googleapis.com/v1/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("userinfo request failed: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		http.Error(w, fmt.Sprintf("userinfo error %d: %s", resp.StatusCode, string(b)), http.StatusBadGateway)
		return
	}

	// Tagasta JSON (kiire kontrolliks). Hiljem saad siit luua/uuendada local user’i ja sessiooni.
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(json.RawMessage(b))
}
