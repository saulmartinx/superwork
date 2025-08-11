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
		ClientID:     "694142574846-vje17tjf4dcm6t6v0835l9akld4r8ccm.apps.googleusercontent.com", // uus Client ID
		ClientSecret: "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", // uus Client Secret
		RedirectURL:  config.GoogleRedirect,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}
}
