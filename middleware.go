package main

import (
	"log"
	"net/http"
)

type userRequiredFunc func(w http.ResponseWriter, r *http.Request, user *User)

func requireUser(f userRequiredFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, sessionName)

		userID, exists := session.Values["user_id"].(string)
		if !exists {
			http.Error(w, "User not logged in", http.StatusUnauthorized)
			return
		}

		user, err := selectUserByID(userID)
		if err != nil {
			log.Println(err)
			http.Error(w, "Failed to load user", http.StatusInternalServerError)
			return
		}

		f(w, r, user)
	}
}
