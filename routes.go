package main

import (
	"net/http"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/didip/tollbooth"
	"github.com/gorilla/mux"
)

func defineRoutes() *mux.Router {
	r := mux.NewRouter()

	limit := tollbooth.LimitFuncHandler

	{
		l := tollbooth.NewLimiter(5, time.Second)

		// Oauth
		r.Handle("/api/facebook_login_url", limit(l, handleGetFacebookLoginURL)).Methods("GET")
		r.Handle("/api/google_login_url", limit(l, handleGetGoogleLoginURL)).Methods("GET")
		r.Handle("/api/oauth2callback/google", limit(l, handleGetOauthCallbackGoogle)).Methods("GET")
		r.Handle("/api/oauth2callback/facebook", limit(l, handleGetOauthCallbackFacebook)).Methods("GET")

		r.Handle("/api/logout", limit(l, handleGetLogout)).Methods("GET")

		r.Handle("/api/stats", limit(l, handleStats))

		// My profile/current user
		r.Handle("/api/me", limit(l, handlePostMe)).Methods("POST")
		r.Handle("/api/me", limit(l, requireUser(handleGetMe))).Methods("GET")
		r.Handle("/api/me", limit(l, requireUser(handlePutMe))).Methods("PUT")

		r.Handle("/api/activate/{id}", limit(l, handlePostActivation)).Methods("POST")
		r.Handle("/api/activate/{id}", limit(l, handleGetActivation)).Methods("GET")

		r.Handle("/api/token_login/{api_token}", limit(l, handleGetTokenLogin)).Methods("GET")

		r.Handle("/api/time_entries", limit(l, requireUser(handleGetTimeEntries))).Methods("GET")
		r.Handle("/api/time_entries", limit(l, requireUser(handlePostTimeEntries))).Methods("POST")
		r.Handle("/api/time_entries/{id}", limit(l, requireUser(handlePutTimeEntry))).Methods("PUT")
		r.Handle("/api/time_entries/{id}", limit(l, requireUser(handleDeleteTimeEntry))).Methods("DELETE")
		r.Handle("/api/time_entries/{id}", limit(l, requireUser(handleGetTimeEntry))).Methods("GET")

		r.Handle("/api/user_events", limit(l, requireUser(handleGetUserEvents))).Methods("GET")
		r.Handle("/api/user_events", limit(l, requireUser(handlePostUserEvents))).Methods("POST")

		r.Handle("/api/company_users", limit(l, requireUser(handleGetCompanyUsers))).Methods("GET")
		r.Handle("/api/company_users", limit(l, requireUser(handlePostCompanyUsers))).Methods("POST")
		r.Handle("/api/company_users/{id}", limit(l, requireUser(handlePutCompanyUser))).Methods("PUT")
		r.Handle("/api/company_users/{id}", limit(l, requireUser(handleDeleteCompanyUser))).Methods("DELETE")

		r.Handle("/api/companies", limit(l, requireUser(handlePostCompanies))).Methods("POST")
		r.Handle("/api/companies", limit(l, requireUser(handleGetCompanies))).Methods("GET")
		r.Handle("/api/companies/{id}", limit(l, requireUser(handlePutCompany))).Methods("PUT")
		r.Handle("/api/companies/{id}", limit(l, requireUser(handleDeleteCompany))).Methods("DELETE")

		r.Handle("/api/deleted_objects", limit(l, requireUser(handleGetDeletedObjects))).Methods("GET")
		r.Handle("/api/deleted_objects/{id}", limit(l, requireUser(handleUndeletedObject))).Methods("DELETE")

		r.Handle("/api/activities", limit(l, requireUser(handleGetActivities))).Methods("GET")
		r.Handle("/api/activities", limit(l, requireUser(handlePostActivities))).Methods("POST")
		r.Handle("/api/activities/{id}", limit(l, requireUser(handlePutActivity))).Methods("PUT")
		r.Handle("/api/activities/{id}", limit(l, requireUser(handleDeleteActivity))).Methods("DELETE")

		r.Handle("/api/activity_fields", limit(l, requireUser(handleGetActivityFields))).Methods("GET")
		r.Handle("/api/activity_fields", limit(l, requireUser(handlePostActivityFields))).Methods("POST")
		r.Handle("/api/activity_fields/{id}", limit(l, requireUser(handlePutActivityField))).Methods("PUT")
		r.Handle("/api/activity_fields/{id}", limit(l, requireUser(handleDeleteActivityField))).Methods("DELETE")

		r.Handle("/api/activity_types", limit(l, requireUser(handleGetActivityTypes))).Methods("GET")
		r.Handle("/api/activity_types", limit(l, requireUser(handlePostActivityTypes))).Methods("POST")
		r.Handle("/api/activity_types/{id}", limit(l, requireUser(handlePutActivityType))).Methods("PUT")
		r.Handle("/api/activity_types/{id}", limit(l, requireUser(handleDeleteActivityType))).Methods("DELETE")

		r.Handle("/api/currencies", limit(l, requireUser(handleGetCurrencies))).Methods("GET")
		r.Handle("/api/currencies", limit(l, requireUser(handlePostCurrencies))).Methods("POST")
		r.Handle("/api/currencies/{id}", limit(l, requireUser(handlePutCurrency))).Methods("PUT")
		r.Handle("/api/currencies/{id}", limit(l, requireUser(handleDeleteCurrency))).Methods("DELETE")

		r.Handle("/api/tasks", limit(l, requireUser(handleGetTasks))).Methods("GET")
		r.Handle("/api/tasks", limit(l, requireUser(handlePostTasks))).Methods("POST")
		r.Handle("/api/tasks/{id}", limit(l, requireUser(handleGetTask))).Methods("GET")
		r.Handle("/api/tasks/{id}", limit(l, requireUser(handlePutTask))).Methods("PUT")
		r.Handle("/api/tasks/{id}", limit(l, requireUser(handleDeleteTask))).Methods("DELETE")

		r.Handle("/api/task_fields", limit(l, requireUser(handleGetTaskFields))).Methods("GET")
		r.Handle("/api/task_fields", limit(l, requireUser(handlePostTaskFields))).Methods("POST")
		r.Handle("/api/task_fields/{id}", limit(l, requireUser(handlePutTaskField))).Methods("PUT")
		r.Handle("/api/task_fields/{id}", limit(l, requireUser(handleDeleteTaskField))).Methods("DELETE")

		r.Handle("/api/files", limit(l, requireUser(handleGetFiles))).Methods("GET")
		r.Handle("/api/files", limit(l, requireUser(handlePostFiles))).Methods("POST")
		r.Handle("/api/files/{id}", limit(l, requireUser(handlePutFile))).Methods("PUT")
		r.Handle("/api/files/{id}", limit(l, requireUser(handleDeleteFile))).Methods("DELETE")

		r.Handle("/api/filters", limit(l, requireUser(handleGetFilters))).Methods("GET")
		r.Handle("/api/filters", limit(l, requireUser(handlePostFilters))).Methods("POST")
		r.Handle("/api/filters/{id}", limit(l, requireUser(handlePutFilter))).Methods("PUT")
		r.Handle("/api/filters/{id}", limit(l, requireUser(handleDeleteFilter))).Methods("DELETE")

		r.Handle("/api/goals", limit(l, requireUser(handleGetGoals))).Methods("GET")
		r.Handle("/api/goals", limit(l, requireUser(handlePostGoals))).Methods("POST")
		r.Handle("/api/goals/{id}", limit(l, requireUser(handlePutGoal))).Methods("PUT")
		r.Handle("/api/goals/{id}", limit(l, requireUser(handleDeleteGoal))).Methods("DELETE")

		r.Handle("/api/notes", limit(l, requireUser(handleGetNotes))).Methods("GET")
		r.Handle("/api/notes", limit(l, requireUser(handlePostNotes))).Methods("POST")
		r.Handle("/api/notes/{id}", limit(l, requireUser(handlePutNote))).Methods("PUT")
		r.Handle("/api/notes/{id}", limit(l, requireUser(handleDeleteNote))).Methods("DELETE")

		r.Handle("/api/note_fields", limit(l, requireUser(handleGetNoteFields))).Methods("GET")
		r.Handle("/api/note_fields", limit(l, requireUser(handlePostNoteFields))).Methods("POST")
		r.Handle("/api/note_fields/{id}", limit(l, requireUser(handlePutNoteField))).Methods("PUT")
		r.Handle("/api/note_fields/{id}", limit(l, requireUser(handleDeleteNoteField))).Methods("DELETE")

		r.Handle("/api/categories", limit(l, requireUser(handleGetCategories))).Methods("GET")
		r.Handle("/api/categories", limit(l, requireUser(handlePostCategories))).Methods("POST")
		r.Handle("/api/categories/{id}", limit(l, requireUser(handlePutCategory))).Methods("PUT")
		r.Handle("/api/categories/{id}", limit(l, requireUser(handleDeleteCategory))).Methods("DELETE")

		r.Handle("/api/organizations", limit(l, requireUser(handleGetOrganizations))).Methods("GET")
		r.Handle("/api/organizations", limit(l, requireUser(handlePostOrganizations))).Methods("POST")
		r.Handle("/api/organizations/{id}", limit(l, requireUser(handlePutOrganization))).Methods("PUT")
		r.Handle("/api/organizations/{id}", limit(l, requireUser(handleDeleteOrganization))).Methods("DELETE")

		r.Handle("/api/organizations_fields", limit(l, requireUser(handleGetOrganizationFields))).Methods("GET")
		r.Handle("/api/organizations_fields", limit(l, requireUser(handlePostOrganizationFields))).Methods("POST")
		r.Handle("/api/organizations_fields/{id}", limit(l, requireUser(handlePutOrganizationField))).Methods("PUT")
		r.Handle("/api/organizations_fields/{id}", limit(l, requireUser(handleDeleteOrganizationField))).Methods("DELETE")

		r.Handle("/api/organizations_relationships", limit(l, requireUser(handleGetOrganizationRelationships))).Methods("GET")
		r.Handle("/api/organizations_relationships", limit(l, requireUser(handlePostOrganizationRelationships))).Methods("POST")
		r.Handle("/api/organizations_relationships/{id}", limit(l, requireUser(handlePutOrganizationRelationship))).Methods("PUT")
		r.Handle("/api/organizations_relationships/{id}", limit(l, requireUser(handleDeleteOrganizationRelationship))).Methods("DELETE")

		r.Handle("/api/contacts", limit(l, requireUser(handleGetContacts))).Methods("GET")
		r.Handle("/api/contacts", limit(l, requireUser(handlePostContacts))).Methods("POST")
		r.Handle("/api/contacts/{id}", limit(l, requireUser(handlePutContact))).Methods("PUT")
		r.Handle("/api/contacts/{id}", limit(l, requireUser(handleDeleteContact))).Methods("DELETE")

		r.Handle("/api/persons/{id}", limit(l, requireUser(handleGetPerson))).Methods("GET")
		r.Handle("/api/persons", limit(l, requireUser(handleGetPersons))).Methods("GET")
		r.Handle("/api/persons", limit(l, requireUser(handlePostPersons))).Methods("POST")
		r.Handle("/api/persons/{id}", limit(l, requireUser(handlePutPerson))).Methods("PUT")
		r.Handle("/api/persons/{id}", limit(l, requireUser(handleDeletePerson))).Methods("DELETE")

		r.Handle("/api/person_fields", limit(l, requireUser(handleGetPersonFields))).Methods("GET")
		r.Handle("/api/person_fields", limit(l, requireUser(handlePostPersonFields))).Methods("POST")
		r.Handle("/api/person_fields/{id}", limit(l, requireUser(handlePutPersonField))).Methods("PUT")
		r.Handle("/api/person_fields/{id}", limit(l, requireUser(handleDeletePersonField))).Methods("DELETE")

		r.Handle("/api/workflows", limit(l, requireUser(handleGetWorkflows))).Methods("GET")
		r.Handle("/api/workflows", limit(l, requireUser(handlePostWorkflows))).Methods("POST")
		r.Handle("/api/workflows/{id}", limit(l, requireUser(handlePutWorkflow))).Methods("PUT")
		r.Handle("/api/workflows/{id}", limit(l, requireUser(handleDeleteWorkflow))).Methods("DELETE")

		r.Handle("/api/prices", limit(l, requireUser(handleGetPrices))).Methods("GET")
		r.Handle("/api/prices", limit(l, requireUser(handlePostPrices))).Methods("POST")
		r.Handle("/api/prices/{id}", limit(l, requireUser(handlePutPrice))).Methods("PUT")
		r.Handle("/api/prices/{id}", limit(l, requireUser(handleDeletePrice))).Methods("DELETE")

		r.Handle("/api/products", limit(l, requireUser(handleGetProducts))).Methods("GET")
		r.Handle("/api/products", limit(l, requireUser(handlePostProducts))).Methods("POST")
		r.Handle("/api/products/{id}", limit(l, requireUser(handlePutProduct))).Methods("PUT")
		r.Handle("/api/products/{id}", limit(l, requireUser(handleDeleteProduct))).Methods("DELETE")

		r.Handle("/api/product_fields", limit(l, requireUser(handleGetProductFields))).Methods("GET")
		r.Handle("/api/product_fields", limit(l, requireUser(handlePostProductFields))).Methods("POST")
		r.Handle("/api/product_fields/{id}", limit(l, requireUser(handlePutProductField))).Methods("PUT")
		r.Handle("/api/product_fields/{id}", limit(l, requireUser(handleDeleteProductField))).Methods("DELETE")

		r.Handle("/api/push_notifications", limit(l, requireUser(handleGetPushNotifications))).Methods("GET")
		r.Handle("/api/push_notifications", limit(l, requireUser(handlePostPushNotifications))).Methods("POST")
		r.Handle("/api/push_notifications/{id}", limit(l, requireUser(handlePutPushNotification))).Methods("PUT")
		r.Handle("/api/push_notifications/{id}", limit(l, requireUser(handleDeletePushNotification))).Methods("DELETE")

		r.Handle("/api/stages", limit(l, requireUser(handleGetStages))).Methods("GET")
		r.Handle("/api/stages", limit(l, requireUser(handlePostStages))).Methods("POST")
		r.Handle("/api/stages/{id}", limit(l, requireUser(handlePutStage))).Methods("PUT")
		r.Handle("/api/stages/{id}", limit(l, requireUser(handleDeleteStage))).Methods("DELETE")

		r.Handle("/api/timeline", limit(l, requireUser(handleGetTimeline))).Methods("GET")
	}

	{
		l := tollbooth.NewLimiter(10, time.Second)

		r.PathPrefix("/").Handler(tollbooth.LimitHandler(l, gziphandler.GzipHandler(http.FileServer(http.Dir(config.Public)))))
	}

	return r
}
