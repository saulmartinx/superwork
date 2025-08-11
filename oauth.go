package main

import (
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"
)

func facebookOauthConf() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     "222679544825854",
		ClientSecret: "c5560b28c59b7d5afdec1b4d07e51a02",
		RedirectURL:  config.FacebookRedirect,
		Scopes:       []string{"email"},
		Endpoint:     facebook.Endpoint,
	}
}

func googleOauthConf() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("SUPERWORK_GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("SUPERWORK_GOOGLE_CLIENT_SECRET"),
		RedirectURL:  config.GoogleRedirect, // https://superwork-o026.onrender.com/api/oauth2callback/google
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
			"openid",
		},
		Endpoint: google.Endpoint,
	}
}
