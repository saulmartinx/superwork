package main

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"
)

func googleOauthConf() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     config.GoogleClientID,
		ClientSecret: config.GoogleClientSecret,
		RedirectURL:  config.GoogleRedirect,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}
}

func facebookOauthConf() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     "", // pole vaja; jätame tühjaks kui FB loginit ei kasuta
		ClientSecret: "",
		RedirectURL:  config.FacebookRedirect,
		Scopes:       []string{"email"},
		Endpoint:     facebook.Endpoint,
	}
}
