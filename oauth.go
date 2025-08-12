package main

import (
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
		ClientID:     "116806776326-1l7bm6htqeof2ftl72339j7d49jped0q.apps.googleusercontent.com",
		ClientSecret: "uMqr7xxUcgY3MJ7_4Ns8YLvB",
		RedirectURL:  config.GoogleRedirect,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}
}
