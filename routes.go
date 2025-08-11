package main

import (
	"net/http"
	"time"

	limiter "github.com/didip/tollbooth/limiter"
	"github.com/NYTimes/gziphandler"
	"github.com/didip/tollbooth"
	"github.com/gorilla/mux"
)

func defineRoutes() *mux.Router {
	r := mux.NewRouter()

	// Rate limit: 1 req/sec, aegumise seade uue tollbooth API jaoks
	l := tollbooth.NewLimiter(1, &limiter.ExpirableOptions{
		DefaultExpirationTTL: time.Second,
	})

	// Mugav mähis: võtab handlerFunc'i ja tagastab limiteriga handleri
	limit := func(h http.HandlerFunc) http.Handler {
		return tollbooth.LimitFuncHandler(l, h)
	}

	{
		// Oauth
		r.Handle("/api/facebook_login_url", limit(handleGetFacebookLoginURL)).Methods("GET")
		r.Handle("/api/google_login_url", limit(handleGetGoogleLoginURL)).Methods("GET")
		r.Handle("/api/oauth2callback/google", limit(handleGetOauthCallbackGoogle)).Methods("GET")
		r.Handle("/api/oauth2callback/facebook", limit(handleGetOauthCallbackFacebook)).Methods("GET")

		r.Handle("/api/logout", limit(handleGetLogout)).Methods("GET")

		r.Handle("/api/stats", limit(handleStats))

		// My profile/current user
		r.Handle("/api/me", limit(handlePostMe)).Methods("POST")
		r.Handle("/api/me", limit(requireUser(handleGetMe))).Methods("GET")
		r.Handle("/api/me", limit(requireUser(handlePutMe))).Methods("PUT")

		r.Handle("/api/activate/{id}", limit(handlePostActivation)).Methods("POST")
		r.Handle("/api/activate/{id}", limit(handleGetActivation)).Methods("GET")

		r.Handle("/api/token_login/{api_token}", limit(handleGetTokenLogin)).Methods("GET")

		r.Handle("/api/time_entries", limit(requireUser(handleGetTimeEntries))).Methods("GET")
		r.Handle("/api/time_entries", limit(requireUser(handlePostTimeEntries))).Methods("POST")
		r.Handle("/api/time_entries/{id}", limit(requireUser(handlePutTimeEntry))).Methods("PUT")
		r.Handle("/api/time_entries/{id}", limit(requireUser(handleDeleteTimeEntry))).Methods("DELETE")
		r.Handle("/api/time_entries/{id}", limit(requireUser(handleGetTimeEntry))).Methods("GET")

		r.Handle("/api/user_events", limit(requireUser(handleGetUserEvents))).Methods("GET")
		r.Handle("/api/user_events", limit(requireUser(handlePostUserEvents))).Methods("POST")

		r.Handle("/api/company_users", limit(requireUser(handleGetCompanyUsers))).Methods("GET")
		r.Handle("/api/company_users", limit(requireUser(handlePostCompanyUsers))).Methods("POST")
		r.Handle("/api/company_users/{id}", limit(requireUser(handlePutCompanyUser))).Methods("PUT")
		r.Handle("/api/company_users/{id}", limit(requireUser(handleDeleteCompanyUser))).Methods("DELETE")

		r.Handle("/api/companies", limit(requireUser(handlePostCompanies))).Methods("POST")
		r.Handle("/api/companies", limit(requireUser(handleGetCompanies))).Methods("GET")
		r.Handle("/api/companies/{id}", limit(requireUser(handlePutCompany))).Methods("PUT")
		r.Handle("/api/companies/{id}", limit(requireUser(handleDeleteCompany))).Methods("DELETE")

		r.Handle("/api/deleted_objects", limit(requireUser(handleGetDeletedObjects))).Methods("GET")
		r.Handle("/api/deleted_objects/{id}", limit(requireUser(handleUndeletedObject))).Methods("DELETE")

		r.Handle("/api/activities", limit(requireUser(handleGetActivities))).Methods("GET")
		r.Handle("/api/activities", limit(requireUser(handlePostActivities))).Methods("POST")
		r.Handle("/api/activities/{id}", limit(requireUser(handlePutActivity))).Methods("PUT")
		r.Handle("/api/activities/{id}", limit(requireUser(handleDeleteActivity))).Methods("DELETE")

		r.Handle("/api/activity_fields", limit(requireUser(handleGetActivityFields))).Methods("GET")
		r.Handle("/api/activity_fields", limit(requireUser(handlePostActivityFields))).Methods("POST")
		r.Handle("/api/activity_fields/{id}", limit(requireUser(handlePutActivityField))).Methods("PUT")
		r.Handle("/api/activity_fields/{id}", limit(requireUser(handleDeleteActivityField))).Methods("DELETE")

		r.Handle("/api/activity_types", limit(requireUser(handleGetActivityTypes))).Methods("GET")
		r.Handle("/api/activity_types", limit(requireUser(handlePostActivityTypes))).Methods("POST")
		r.Handle("/api/activity_types/{id}", limit(requireUser(handlePutActivityType))).Methods("PUT")
		r.Handle("/api/activity_types/{id}", limit(requireUser(handleDeleteActivityType))).Methods("DELETE")

		r.Handle("/api/currencies", limit(requireUser(handleGetCurrencies))).Methods("GET")
		r.Handle("/api/currencies", limit(requireUser(handlePostCurrencies))).Methods("POST")
		r.Handle("/api/currencies/{id}", limit(requireUser(handlePutCurrency))).Methods("PUT")
		r.Handle("/api/currencies/{id}", limit(requireUser(handleDeleteCurrency))).Methods("DELETE")

		r.Handle("/api/tasks", limit(requireUser(handleGetTasks))).Methods("GET")
		r.Handle("/api/tasks", limit(requireUser(handlePostTasks))).Methods("POST")
		r.Handle("/api/tasks/{id}", limit(requireUser(handleGetTask))).Methods("GET")
		r.Handle("/api/tasks/{id}", limit(requireUser(handlePutTask))).Methods("PUT")
		r.Handle("/api/tasks/{id}", limit(requireUser(handleDeleteTask))).Methods("DELETE")

		r.Handle("/api/task_fields", limit(requireUser(handleGetTaskFields))).Methods("GET")
		r.Handle("/api/task_fields", limit(requireUser(handlePostTaskFields))).Methods("POST")
		r.Handle("/api/task_fields/{id}", limit(requireUser(handlePutTaskField))).Methods("PUT")
		r.Handle("/api/task_fields/{id}", limit(requireUser(handleDeleteTaskField))).Methods("DELETE")

		r.Handle("/api/files", limit(requireUser(handleGetFiles))).Methods("GET")
		r.Handle("/api/files", limit(requireUser(handlePostFiles))).Methods("POST")
		r.Handle("/api/files/{id}", limit(requireUser(handlePutFile))).Methods("PUT")
		r.Handle("/api/files/{id}", limit(requireUser(handleDeleteFile))).Methods("DELETE")

		r.Handle("/api/filters", limit(requireUser(handleGetFilters))).Methods("GET")
		r.Handle("/api/filters", limit(requireUser(handlePostFilters))).Methods("POST")
		r.Handle("/api/filters/{id}", limit(requireUser(handlePutFilter))).Methods("PUT")
		r.Handle("/api/filters/{id}", limit(requireUser(handleDeleteFilter))).Methods("DELETE")

		r.Handle("/api/goals", limit(requireUser(handleGetGoals))).Methods("GET")
		r.Handle("/api/goals", limit(requireUser(handlePostGoals))).Methods("POST")
		r.Handle("/api/goals/{id}", limit(requireUser(handlePutGoal))).Methods("PUT")
		r.Handle("/api/goals/{id}", limit(requireUser(handleDeleteGoal))).Methods("DELETE")

		r.Handle("/api/notes", limit(requireUser(handleGetNotes))).Methods("GET")
		r.Handle("/api/notes", limit(requireUser(handlePostNotes))).Methods("POST")
		r.Handle("/api/notes/{id}", limit(requireUser(handlePutNote))).Methods("PUT")
		r.Handle("/api/notes/{id}", limit(requireUser(handleDeleteNote))).Methods("DELETE")

		r.Handle("/api/note_fields", limit(requireUser(handleGetNoteFields))).Methods("GET")
		r.Handle("/api/note_fields", limit(requireUser(handlePostNoteFields))).Methods("POST")
		r.Handle("/api/note_fields/{id}", limit(requireUser(handlePutNoteField))).Methods("PUT")
		r.Handle("/api/note_fields/{id}", limit(requireUser(handleDeleteNoteField))).Methods("DELETE")

		r.Handle("/api/categories", limit(requireUser(handleGetCategories))).Methods("GET")
		r.Handle("/api/categories", limit(requireUser(handlePostCategories))).Methods("POST")
		r.Handle("/api/categories/{id}", limit(requireUser(handlePutCategory))).Methods("PUT")
		r.Handle("/api/categories/{id}", limit(requireUser(handleDeleteCategory))).Methods("DELETE")

		r.Handle("/api/organizations", limit(requireUser(handleGetOrganizations))).Methods("GET")
		r.Handle("/api/organizations", limit(requireUser(handlePostOrganizations))).Methods("POST")
		r.Handle("/api/organizations/{id}", limit(requireUser(handlePutOrganization))).Methods("PUT")
		r.Handle("/api/organizations/{id}", limit(requireUser(handleDeleteOrganization))).Methods("DELETE")

		r.Handle("/api/organizations_fields", limit(requireUser(handleGetOrganizationFields))).Methods("GET")
		r.Handle("/api/organizations_fields", limit(requireUser(handlePostOrganizationFields))).Methods("POST")
		r.Handle("/api/organizations_fields/{id}", limit(requireUser(handlePutOrganizationField))).Methods("PUT")
		r.Handle("/api/organizations_fields/{id}", limit(requireUser(handleDeleteOrganizationField))).Methods("DELETE")

		r.Handle("/api/organizations_relationships", limit(requireUser(handleGetOrganizationRelationships))).Methods("GET")
		r.Handle("/api/organizations_relationships", limit(requireUser(handlePostOrganizationRelationships))).Methods("POST")
		r.Handle("/api/organizations_relationships/{id}", limit(requireUser(handlePutOrganizationRelationship))).Methods("PUT")
		r.Handle("/api/organizations_relationships/{id}", limit(requireUser(handleDeleteOrganizationRelationship))).Methods("DELETE")

		r.Handle("/api/contacts", limit(requireUser(handleGetContacts))).Methods("GET")
		r.Handle("/api/contacts", limit(requireUser(handlePostContacts))).Methods("POST")
		r.Handle("/api/contacts/{id}", limit(requireUser(handlePutContact))).Methods("PUT")
		r.Handle("/api/contacts/{id}", limit(requireUser(handleDeleteContact))).Methods("DELETE")

		r.Handle("/api/persons/{id}", limit(requireUser(handleGetPerson))).Methods("GET")
		r.Handle("/api/persons", limit(requireUser(handleGetPersons))).Methods("GET")
		r.Handle("/api/persons", limit(requireUser(handlePostPersons))).Methods("POST")
		r.Handle("/api/persons/{id}", limit(requireUser(handlePutPerson))).Methods("PUT")
		r.Handle("/api/persons/{id}", limit(requireUser(handleDeletePerson))).Methods("DELETE")

		r.Handle("/api/person_fields", limit(requireUser(handleGetPersonFields))).Methods("GET")
		r.Handle("/api/person_fields", limit(requireUser(handlePostPersonFields))).Methods("POST")
		r.Handle("/api/person_fields/{id}", limit(requireUser(handlePutPersonField))).Methods("PUT")
		r.Handle("/api/person_fields/{id}", limit(requireUser(handleDeletePersonField))).Methods("DELETE")

		r.Handle("/api/workflows", limit(requireUser(handleGetWorkflows))).Methods("GET")
		r.Handle("/api/workflows", limit(requireUser(handlePostWorkflows))).Methods("POST")
		r.Handle("/api/workflows/{id}", limit(requireUser(handlePutWorkflow))).Methods("PUT")
		r.Handle("/api/workflows/{id}", limit(requireUser(handleDeleteWorkflow))).Methods("DELETE")

		r.Handle("/api/prices", limit(requireUser(handleGetPrices))).Methods("GET")
		r.Handle("/api/prices", limit(requireUser(handlePostPrices))).Methods("POST")
		r.Handle("/api/prices/{id}", limit(requireUser(handlePutPrice))).Methods("PUT")
		r.Handle("/api/prices/{id}", limit(requireUser(handleDeletePrice))).Methods("DELETE")

		r.Handle("/api/products", limit(requireUser(handleGetProducts))).Methods("GET")
		r.Handle("/api/products", limit(requireUser(handlePostProducts))).Methods("POST")
		r.Handle("/api/products/{id}", limit(requireUser(handlePutProduct))).Methods("PUT")
		r.Handle("/api/products/{id}", limit(requireUser(handleDeleteProduct))).Methods("DELETE")

		r.Handle("/api/product_fields", limit(requireUser(handleGetProductFields))).Methods("GET")
		r.Handle("/api/product_fields", limit(requireUser(handlePostProductFields))).Methods("POST")
		r.Handle("/api/product_fields/{id}", limit(requireUser(handlePutProductField))).Methods("PUT")
		r.Handle("/api/product_fields/{id}", limit(requireUser(handleDeleteProductField))).Methods("DELETE")

		r.Handle("/api/push_notifications", limit(requireUser(handleGetPushNotifications))).Methods("GET")
		r.Handle("/api/push_notifications", limit(requireUser(handlePostPushNotifications))).Methods("POST")
		r.Handle("/api/push_notifications/{id}", limit(requireUser(handlePutPushNotification))).Methods("PUT")
		r.Handle("/api/push_notifications/{id}", limit(requireUser(handleDeletePushNotification))).Methods("DELETE")

		r.Handle("/api/stages", limit(requireUser(handleGetStages))).Methods("GET")
		r.Handle("/api/stages", limit(requireUser(handlePostStages))).Methods("POST")
		r.Handle("/api/stages/{id}", limit(requireUser(handlePutStage))).Methods("PUT")
		r.Handle("/api/stages/{id}", limit(requireUser(handleDeleteStage))).Methods("DELETE")

		r.Handle("/api/timeline", limit(requireUser(handleGetTimeline))).Methods("GET")
	}

	{
		// Static assets with gzip + rate limit
		r.PathPrefix("/").Handler(
			tollbooth.LimitHandler(
				l,
				gziphandler.GzipHandler(http.FileServer(http.Dir(config.Public))),
			),
		)
	}

	return r
}
