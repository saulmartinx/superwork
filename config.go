package main

// Config store application configuration
type Config struct {
	Env           string
	GeocodeAPIKey string `envconfig:"geocode_api_key" default:"AIzaSyBXEzuZ_6v_rUKy4OFhhkla1f7qKL7mqsU"`
	BugsnagAPIKey string `envconfig:"bugsnag_api_key" default:"86e4cd618565b8b302bd8dc13574c3d4"`
	AdminEmail    string `envconfig:"admin_email" default:"support@superwork.io"`
	ZohoPassword  string `envconfig:"zoho_password"`
	Port          int    `default:"8000"`
	Public        string `default:"public"`
	// Folder where app is located (app root folder)
	Dir              string `default:""`
	Logfile          string `default:"superwork.log"`
	Log              bool   `default:"false"`
	Secret           string `default:"z8a0YXYgDwmyDW0USaHBC4CmnUMU5QbwYXPRLNoHf5LmOMJvSbfVWuHAiPDVFE7"`
	GoogleRedirect   string `envconfig:"google_redirect" default:"http://localhost:8000/api/oauth2callback/google"`
	FacebookRedirect string `envconfig:"facebook_redirect" default:"http://localhost.superwork.io:8000/api/oauth2callback/facebook"`
}

var config Config
