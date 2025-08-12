package main

import (
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"
)

func googleOauthConf() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("SUPERWORK_GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("SUPERWORK_GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("SUPERWORK_GOOGLE_REDIRECT"),
		Scopes: []string{
            "openid",
            "https://www.googleapis.com/auth/userinfo.profile",
            "https://www.googleapis.com/auth/userinfo.email",
},

		},
		Endpoint: google.Endpoint,
	}
}

func facebookOauthConf() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("SUPERWORK_FACEBOOK_CLIENT_ID"),
		ClientSecret: os.Getenv("SUPERWORK_FACEBOOK_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("SUPERWORK_FACEBOOK_REDIRECT"),
		Scopes:       []string{"email"},
		Endpoint:     facebook.Endpoint,
	}
}

