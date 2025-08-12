package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)

func handleGetFacebookLoginURL(w http.ResponseWriter, r *http.Request) {
	state := uuid.NewV4().String()
	session, _ := store.Get(r, sessionName)
	session.Values["login_state"] = state
	session.Save(r, w)

	url := facebookOauthConf().AuthCodeURL(state)
	w.Write([]byte(url))
}

func handleGetGoogleLoginURL(w http.ResponseWriter, r *http.Request) {
	state := uuid.NewV4().String()
	session, _ := store.Get(r, sessionName)
	session.Values["login_state"] = state
	session.Save(r, w)

	url := googleOauthConf().AuthCodeURL(state)
	w.Write([]byte(url))
}

func handleGetOauthCallback(url string, conf *oauth2.Config, w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	session, _ := store.Get(r, sessionName)
	if session.Values["login_state"] != state {
		log.Println("Invalid oauth login state, expected", session.Values["login_state"], "but got", state)
		http.Error(w, "Invalid oauth login state", http.StatusBadRequest)
		return
	}
	session.Values["login_state"] = nil
	session.Save(r, w)

	tok, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error getting oauth login token", http.StatusInternalServerError)
		return
	}

	log.Println("getting oauth data from", url)

	client := conf.Client(oauth2.NoContext, tok)
	resp, err := client.Get(url)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error getting oauth user data", http.StatusInternalServerError)
		return
	}

	var profile map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		log.Println(err)
		http.Error(w, "Error reading oauth user data", http.StatusInternalServerError)
		return
	}

	log.Println("oauth response", profile)

	email := profile["email"].(string)
	name := profile["name"].(string)
	picture := ""
	if s, isString := profile["picture"].(string); isString {
		// Google
		picture = s
	} else if pic, isMap := profile["picture"].(map[string]interface{}); isMap {
		// Facebook
		if data, isMap := pic["data"].(map[string]interface{}); isMap {
			if s, isString = data["url"].(string); isString {
				picture = s
			}
		}
	}
	user, err := selectUserByEmail(email)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error loading user data", http.StatusInternalServerError)
		return
	}

	if nil == user {
		user = &User{}
		user.Email = email
		user.Name = name
		user.Picture = &picture

		company := Company{}
		if user.Name == "" {
			company.Name = user.Email + " company"
		} else {
			company.Name = user.Name + " company"
		}
		if err := insertCompany(&company); err != nil {
			log.Println(err)
			http.Error(w, "Error creating company", http.StatusInternalServerError)
			return
		}

		user.ActiveCompanyID = company.ID
		if err := insertUser(user); err != nil {
			log.Println(err)
			http.Error(w, "Error inserting user data", http.StatusInternalServerError)
			return
		}

		timeline := Timeline{
			UnderCompanyID: company.ID,
			UserID:         user.ID,
			CompanyID:      company.ID,
			Action:         "created",
		}
		timeline.Name = company.Name
		if err := insertTimeline(&timeline); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		companyUser := CompanyUser{
			CompanyID: company.ID,
			UserID:    user.ID,
		}
		if err := insertCompanyUser(&companyUser); err != nil {
			log.Println(err)
			http.Error(w, "Error creating company user", http.StatusInternalServerError)
			return
		}

	} else {
		user.Name = name
		user.Picture = &picture
		if err := updateUser(*user); err != nil {
			log.Println("Error updating user", err)
			http.Error(w, "Error updating user data", http.StatusInternalServerError)
			return
		}
	}

	session.Values["user_id"] = user.ID
	session.Save(r, w)

	if err := populateUser(*user); err != nil {
		log.Println(err)
		http.Error(w, "Error populating user data", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func handleGetOauthCallbackGoogle(w http.ResponseWriter, r *http.Request) {
	handleGetOauthCallback("https://www.googleapis.com/oauth2/v1/userinfo", googleOauthConf(), w, r)
}

func handleGetOauthCallbackFacebook(w http.ResponseWriter, r *http.Request) {
	handleGetOauthCallback("https://graph.facebook.com/me?fields=name,email,picture.type(large)", facebookOauthConf(), w, r)
}

func handleGetLogout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, sessionName)
	session.Values["user_id"] = nil
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound)
}

func handleGetMe(w http.ResponseWriter, r *http.Request, user *User) {
	w.Write(must(json.Marshal(user)))
}

func handlePutMe(w http.ResponseWriter, r *http.Request, user *User) {
	var input User
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	activeCompanyChanged := user.ActiveCompanyID != input.ActiveCompanyID

	user.Phone = input.Phone
	user.YearOfBirth = input.YearOfBirth
	user.ActiveCompanyID = input.ActiveCompanyID
	user.ActiveWorkflowID = input.ActiveWorkflowID

	if activeCompanyChanged {
		var err error
		user.ActiveWorkflowID, user.ActiveWorkflowName, err = selectFirstWorkflowByCompany(user.ActiveCompanyID)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if input.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		user.PasswordHash = string(hashedPassword)
	}

	if err := updateUser(*user); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal(user)))
}

func handlePostMe(w http.ResponseWriter, r *http.Request) {
	var input User
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(input.Email) == 0 {
		http.Error(w, "E-mail cannot be empty", http.StatusBadRequest)
		return
	}
	if len(input.Password) == 0 {
		http.Error(w, "Password cannot be empty", http.StatusBadRequest)
		return
	}

	// Check for existing user

	existingUser, err := selectUserByEmail(input.Email)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error loading user data", http.StatusInternalServerError)
		return
	}
	if existingUser != nil {
		// Check if password matches with the existing user. If yes, just log the user in
		err := bcrypt.CompareHashAndPassword([]byte(existingUser.PasswordHash), []byte(input.Password))
		if err == nil {
			session, _ := store.Get(r, sessionName)
			session.Values["user_id"] = existingUser.ID
			session.Save(r, w)

			if err := populateUser(*existingUser); err != nil {
				log.Println(err)
				http.Error(w, "Error populating user data", http.StatusInternalServerError)
				return
			}

			w.Write(must(json.Marshal(existingUser)))
			return
		}

		http.Error(w, "Invalid e-mail or password", http.StatusBadRequest)
		return
	}

	// New user

	var user User
	user.Email = input.Email
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	company := Company{}
	if user.Name == "" {
		company.Name = user.Email + " company"
	} else {
		company.Name = user.Name + " company"
	}
	if err := insertCompany(&company); err != nil {
		log.Println(err)
		http.Error(w, "Error creating company", http.StatusInternalServerError)
		return
	}
	user.ActiveCompanyID = company.ID

	user.PasswordHash = string(hashedPassword)
	user.HasPassword = true

	if err := insertUser(&user); err != nil {
		log.Println(err)
		http.Error(w, "Error inserting user data", http.StatusInternalServerError)
		return
	}

	timeline := Timeline{
		UnderCompanyID: company.ID,
		UserID:         user.ID,
		CompanyID:      company.ID,
		Action:         "created",
	}
	timeline.Name = company.Name
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	companyUser := CompanyUser{
		CompanyID: company.ID,
		UserID:    user.ID,
	}
	if err := insertCompanyUser(&companyUser); err != nil {
		log.Println(err)
		http.Error(w, "Error creating company user", http.StatusInternalServerError)
		return
	}

	session, _ := store.Get(r, sessionName)
	session.Values["user_id"] = user.ID
	session.Save(r, w)

	if err := populateUser(user); err != nil {
		log.Println(err)
		http.Error(w, "Error populating user data", http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal(user)))
}

func handleGetActivation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	activationID := vars["id"]
	activation, err := selectActivationByID(activationID)
	if err != nil {
		http.Error(w, "Error checking activation code", http.StatusBadRequest)
		return
	}
	if activation == nil {
		http.Error(w, "Activation code is not valid any more", http.StatusBadRequest)
		return
	}
	w.Write(must(json.Marshal("ok")))
}

func handlePostActivation(w http.ResponseWriter, r *http.Request) {
	var input map[string]string
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(input["password"]) == 0 {
		http.Error(w, "Password cannot be empty", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	activationID := vars["id"]
	activation, err := selectActivationByID(activationID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error loading activation", http.StatusInternalServerError)
		return
	}

	user, err := selectUserByID(activation.UserID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error loading user data", http.StatusInternalServerError)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input["password"]), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	user.PasswordHash = string(hashedPassword)
	if err := updateUser(*user); err != nil {
		log.Println(err)
		http.Error(w, "Error inserting user data", http.StatusInternalServerError)
		return
	}

	if err := deleteActivation(activation.ID); err != nil {
		log.Println(err)
		http.Error(w, "Error clearing activation", http.StatusInternalServerError)
		return
	}

	session, _ := store.Get(r, sessionName)
	session.Values["user_id"] = user.ID
	session.Save(r, w)

	w.Write(must(json.Marshal(user)))
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	stats, err := selectStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Error collecting stats", err)
		return
	}
	w.Write(must(json.MarshalIndent(stats, "", "  ")))
}

func handleGetCompanyUsers(w http.ResponseWriter, r *http.Request, user *User) {
	models, err := selectCompanyUsersByCompany(user.ActiveCompanyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.Write(must(json.Marshal(models)))
}

func handlePostCompanyUsers(w http.ResponseWriter, r *http.Request, user *User) {
	var input CompanyUser
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if input.Email == "" {
		http.Error(w, "Email cannot be empty", http.StatusBadRequest)
		return
	}
	if input.Name == "" {
		http.Error(w, "Name cannot be empty", http.StatusBadRequest)
		return
	}

	input.CompanyID = user.ActiveCompanyID

	company, err := selectCompanyByID(user.ActiveCompanyID)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newUser, err := selectUserByEmail(input.Email)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if newUser == nil {
		newUser = &User{}
		newUser.Name = input.Name
		newUser.Email = input.Email
		newUser.ActiveCompanyID = user.ActiveCompanyID
		if err := insertUser(newUser); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// this user has no password and no means to log in (except with FB/Google).
		// so send it an account activation link
		activation := Activation{}
		activation.UserID = newUser.ID
		if err := insertActivation(&activation); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		go sendEmail(
			newUser.Email,
			fmt.Sprintf("%s (%s) added you to company %s.",
				user.Name, user.Email, company.Name),
			fmt.Sprintf("%s (%s) added you to company %s on http://superwork.io. Activate your account by visiting http://superwork.io/#activate/%s",
				user.Name, user.Email, company.Name, activation.ID))
	} else {
		// this user can already log in. just let it know that it was added to the company
		go sendEmail(
			newUser.Email,
			fmt.Sprintf("%s (%s) added you to company %s.",
				user.Name, user.Email, company.Name),
			fmt.Sprintf("%s (%s) added you to company %s on http://superwork.io. Enjoy!",
				user.Name, user.Email, company.Name))
	}

	input.UserID = newUser.ID

	existingCompanyUser, err := selectCompanyUserByUserAndCompany(input.UserID, input.CompanyID)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if existingCompanyUser == nil {
		if err := insertCompanyUser(&input); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		existingCompanyUser = &input
	}

	timeline := Timeline{
		UnderCompanyID: company.ID,
		UserID:         user.ID,
		CompanyUserID:  existingCompanyUser.ID,
		Action:         "created",
	}
	timeline.Name = newUser.Email
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal(existingCompanyUser)))
}

func handlePutCompanyUser(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleDeleteCompanyUser(w http.ResponseWriter, r *http.Request, user *User) {
	vars := mux.Vars(r)
	companyUserID := vars["id"]

	companyUser, err := selectCompanyUserByID(companyUserID)
	if err != nil {
		log.Println(err)
		http.Error(w, "error loading company user", http.StatusInternalServerError)
		return
	}

	existingUser, err := selectUserByID(companyUser.UserID)
	if err != nil {
		log.Println(err)
		http.Error(w, "error loading user", http.StatusInternalServerError)
		return
	}

	if err := deleteCompanyUser(companyUserID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	timeline := Timeline{
		UnderCompanyID: companyUser.CompanyID,
		UserID:         user.ID,
		CompanyUserID:  companyUser.ID,
		Action:         "deleted",
	}
	timeline.Name = existingUser.Email
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal("ok")))
}

func handleGetTimeline(w http.ResponseWriter, r *http.Request, user *User) {
	models, err := selectTimelineByCompany(user.ActiveCompanyID)
	if err != nil {
		http.Error(w, "Error loading timeline", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Write(must(json.Marshal(models)))
}

func handleUndeletedObject(w http.ResponseWriter, r *http.Request, user *User) {
	var input DeletedObject

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := undelete(input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write(must(json.Marshal("ok")))
}

func handleGetDeletedObjects(w http.ResponseWriter, r *http.Request, user *User) {
	models, err := selectDeletedObjectsByCompany(user.ActiveCompanyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Write(must(json.Marshal(models)))
}

func handleGetCompanies(w http.ResponseWriter, r *http.Request, user *User) {
	models, err := selectCompaniesByUser(user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Write(must(json.Marshal(models)))
}

func handlePostCompanies(w http.ResponseWriter, r *http.Request, user *User) {
	var input Company
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := insertCompany(&input); err != nil {
		log.Println(err)
		http.Error(w, "Error creating company", http.StatusInternalServerError)
		return
	}

	timeline := Timeline{
		UnderCompanyID: input.ID,
		UserID:         user.ID,
		CompanyID:      input.ID,
		Action:         "created",
	}
	timeline.Name = input.Name
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	companyUser := CompanyUser{
		CompanyID: input.ID,
		UserID:    user.ID,
	}
	if err := insertCompanyUser(&companyUser); err != nil {
		log.Println(err)
		http.Error(w, "Error creating company", http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal(input)))
}

func handlePutCompany(w http.ResponseWriter, r *http.Request, user *User) {
	var input Company
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := updateCompany(input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	timeline := Timeline{
		UnderCompanyID: user.ActiveCompanyID,
		UserID:         user.ID,
		CompanyID:      input.ID,
		Action:         "updated",
	}
	timeline.Name = input.Name
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	model, err := selectCompanyByID(input.ID)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal(model)))
}

func handleDeleteCompany(w http.ResponseWriter, r *http.Request, user *User) {
	vars := mux.Vars(r)
	ID := vars["id"]

	if err := deleteCompany(ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Write(must(json.Marshal("ok")))
}

func handleGetActivities(w http.ResponseWriter, r *http.Request, user *User) {
	taskID := r.URL.Query().Get("task_id")
	personID := r.URL.Query().Get("person_id")
	orgID := r.URL.Query().Get("org_id")

	var models []Activity
	var err error

	if taskID != "" {
		models, err = selectActivitiesByTask(taskID)
	} else if personID != "" {
		models, err = selectActivitiesByPerson(personID)
	} else if orgID != "" {
		models, err = selectActivitiesByOrganization(orgID)
	} else {
		models, err = selectActivitiesByCompany(user.ActiveCompanyID)
	}

	if err != nil {
		log.Println(err)
		http.Error(w, "Error loading activities", http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal(models)))
}

func handlePostActivities(w http.ResponseWriter, r *http.Request, user *User) {
	var input Activity
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	input.CompanyID = user.ActiveCompanyID
	input.UserID = user.ID
	input.OwnerName = user.Name

	if err := assignOrganization(&input, *user); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := assignPerson(&input, *user); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := insertActivity(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	timeline := Timeline{
		UnderCompanyID: user.ActiveCompanyID,
		UserID:         user.ID,
		ActivityID:     input.ID,
		Action:         "created",
	}
	timeline.Name = input.Name
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	model, err := selectActivityByID(input.ID)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal(model)))
}

func handlePutActivity(w http.ResponseWriter, r *http.Request, user *User) {
	var input Activity
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println(err)
		http.Error(w, "Error parsing task", http.StatusBadRequest)
		return
	}

	if err := assignOrganization(&input, *user); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := assignPerson(&input, *user); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := updateActivity(input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	timeline := Timeline{
		UnderCompanyID: user.ActiveCompanyID,
		UserID:         user.ID,
		ActivityID:     input.ID,
		Action:         "updated",
	}
	timeline.Name = input.Name
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	model, err := selectActivityByID(input.ID)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal(model)))
}

func handleDeleteActivity(w http.ResponseWriter, r *http.Request, user *User) {
	vars := mux.Vars(r)
	ID := vars["id"]

	model, err := selectActivityByID(ID)
	if err != nil {
		log.Println(err)
		http.Error(w, "error loading activity", http.StatusInternalServerError)
		return
	}

	if err := deleteActivity(ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	timeline := Timeline{
		UnderCompanyID: user.ActiveCompanyID,
		UserID:         user.ID,
		ActivityID:     ID,
		Action:         "deleted",
	}
	timeline.Name = model.Name
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal("ok")))
}

func handleGetActivityFields(w http.ResponseWriter, r *http.Request, user *User) {
}

func handlePostActivityFields(w http.ResponseWriter, r *http.Request, user *User) {
}

func handlePutActivityField(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleDeleteActivityField(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleGetActivityTypes(w http.ResponseWriter, r *http.Request, user *User) {
	models, err := selectActivityTypesByCompany(user.ActiveCompanyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Write(must(json.Marshal(models)))
}

func handlePostActivityTypes(w http.ResponseWriter, r *http.Request, user *User) {
}

func handlePutActivityType(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleDeleteActivityType(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleGetCurrencies(w http.ResponseWriter, r *http.Request, user *User) {
	models, err := selectCurrenciesByCompany(user.ActiveCompanyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.Write(must(json.Marshal(models)))
}

func handlePostCurrencies(w http.ResponseWriter, r *http.Request, user *User) {
}

func handlePutCurrency(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleDeleteCurrency(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleGetTasks(w http.ResponseWriter, r *http.Request, user *User) {
	personID := r.URL.Query().Get("person_id")
	orgID := r.URL.Query().Get("org_id")
	onlyActiveTasks := r.URL.Query().Get("only_active_tasks") == "true"

	var models []Task
	var err error

	if personID != "" {
		models, err = selectTasksByPerson(personID)
	} else if orgID != "" {
		models, err = selectTasksByOrganization(orgID)
	} else {
		models, err = selectTasksByWorkflow(user.ActiveWorkflowID, onlyActiveTasks)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Write(must(json.Marshal(models)))
}

func handleGetTask(w http.ResponseWriter, r *http.Request, user *User) {
	vars := mux.Vars(r)
	ID := vars["id"]

	model, err := selectTaskByID(ID)
	if err != nil {
		http.Error(w, "Error loading task", http.StatusBadRequest)
		return
	}

	w.Write(must(json.Marshal(model)))
}

func handlePostTasks(w http.ResponseWriter, r *http.Request, user *User) {
	var input Task
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	input.CompanyID = user.ActiveCompanyID
	input.WorkflowID = user.ActiveWorkflowID
	input.CreatorUserID = user.ID

	if err := assignOrganization(&input, *user); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := assignPerson(&input, *user); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := insertTask(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	timeline := Timeline{
		UnderCompanyID: user.ActiveCompanyID,
		UserID:         user.ID,
		TaskID:         input.ID,
		Action:         "created",
	}
	timeline.Name = input.Name
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	model, err := selectTaskByID(input.ID)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal(model)))
}

func handlePutTask(w http.ResponseWriter, r *http.Request, user *User) {
	var input Task
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println(err)
		http.Error(w, "Error parsing task", http.StatusBadRequest)
		return
	}

	if err := assignOrganization(&input, *user); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := assignPerson(&input, *user); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := updateTask(input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	timeline := Timeline{
		UnderCompanyID: input.CompanyID,
		UserID:         user.ID,
		TaskID:         input.ID,
		Action:         "updated",
	}
	timeline.Name = input.Name
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	model, err := selectTaskByID(input.ID)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal(model)))
}

func handleDeleteTask(w http.ResponseWriter, r *http.Request, user *User) {
	vars := mux.Vars(r)
	ID := vars["id"]

	task, err := selectTaskByID(ID)
	if err != nil {
		log.Println(err)
		http.Error(w, "error loading task", http.StatusInternalServerError)
		return
	}

	if err := deleteTask(ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	timeline := Timeline{
		UnderCompanyID: user.ActiveCompanyID,
		UserID:         user.ID,
		TaskID:         ID,
		Action:         "deleted",
	}
	timeline.Name = task.Name
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal("ok")))
}

func handleGetTaskFields(w http.ResponseWriter, r *http.Request, user *User) {
}

func handlePostTaskFields(w http.ResponseWriter, r *http.Request, user *User) {
}

func handlePutTaskField(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleDeleteTaskField(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleGetFiles(w http.ResponseWriter, r *http.Request, user *User) {
}

func handlePostFiles(w http.ResponseWriter, r *http.Request, user *User) {
}

func handlePutFile(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleDeleteFile(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleGetFilters(w http.ResponseWriter, r *http.Request, user *User) {
}

func handlePostFilters(w http.ResponseWriter, r *http.Request, user *User) {
}

func handlePutFilter(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleDeleteFilter(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleGetGoals(w http.ResponseWriter, r *http.Request, user *User) {
}

func handlePostGoals(w http.ResponseWriter, r *http.Request, user *User) {
}

func handlePutGoal(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleDeleteGoal(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleGetNotes(w http.ResponseWriter, r *http.Request, user *User) {
	taskID := r.URL.Query().Get("task_id")
	personID := r.URL.Query().Get("person_id")
	orgID := r.URL.Query().Get("org_id")

	var models []Note
	var err error

	if taskID != "" {
		models, err = selectNotesByTask(taskID)
	} else if personID != "" {
		models, err = selectNotesByPerson(personID)
	} else if orgID != "" {
		models, err = selectNotesByOrganization(orgID)
	}

	if err != nil {
		log.Println(err)
		http.Error(w, "Error loading notes", http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal(models)))
}

func handlePostNotes(w http.ResponseWriter, r *http.Request, user *User) {
	var input Note
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	input.CompanyID = user.ActiveCompanyID
	input.UserID = user.ID
	input.UserName = user.Name

	if err := insertNote(&input); err != nil {
		log.Println(err)
		http.Error(w, "Error creating note", http.StatusInternalServerError)
		return
	}

	timeline := Timeline{
		UnderCompanyID: input.CompanyID,
		UserID:         user.ID,
		NoteID:         input.ID,
		Action:         "created",
	}
	timeline.Name = input.Name
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal(input)))
}

func handlePutNote(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleDeleteNote(w http.ResponseWriter, r *http.Request, user *User) {
	vars := mux.Vars(r)
	ID := vars["id"]

	note, err := selectNoteByID(ID)
	if err != nil {
		log.Println(err)
		http.Error(w, "error selecting note", http.StatusInternalServerError)
		return
	}

	if err := deleteNote(ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	timeline := Timeline{
		UnderCompanyID: note.CompanyID,
		UserID:         user.ID,
		NoteID:         ID,
		Action:         "deleted",
	}
	timeline.Name = note.Name
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal("ok")))
}

func handleGetNoteFields(w http.ResponseWriter, r *http.Request, user *User) {
}

func handlePostNoteFields(w http.ResponseWriter, r *http.Request, user *User) {
}

func handlePutNoteField(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleDeleteNoteField(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleGetCategories(w http.ResponseWriter, r *http.Request, user *User) {
}

func handlePostCategories(w http.ResponseWriter, r *http.Request, user *User) {
}

func handlePutCategory(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleDeleteCategory(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleGetOrganizations(w http.ResponseWriter, r *http.Request, user *User) {
	models, err := selectOrganizationsByCompany(user.ActiveCompanyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.Write(must(json.Marshal(models)))
}

func handlePostOrganizations(w http.ResponseWriter, r *http.Request, user *User) {
	var input Organization
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	input.CompanyID = user.ActiveCompanyID
	input.OwnerID = user.ID

	if err := insertOrganization(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	timeline := Timeline{
		UnderCompanyID: user.ActiveCompanyID,
		UserID:         user.ID,
		OrganizationID: input.ID,
		Action:         "created",
	}
	timeline.Name = input.Name
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	model, err := selectOrganizationByID(input.ID)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal(model)))
}

func handlePutOrganization(w http.ResponseWriter, r *http.Request, user *User) {
	var input Organization
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := updateOrganization(input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	timeline := Timeline{
		UnderCompanyID: user.ActiveCompanyID,
		UserID:         user.ID,
		OrganizationID: input.ID,
		Action:         "updated",
	}
	timeline.Name = input.Name
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	model, err := selectOrganizationByID(input.ID)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal(model)))
}

func handleDeleteOrganization(w http.ResponseWriter, r *http.Request, user *User) {
	vars := mux.Vars(r)
	ID := vars["id"]

	model, err := selectOrganizationByID(ID)
	if err != nil {
		log.Println(err)
		http.Error(w, "error loading organization", http.StatusInternalServerError)
		return
	}

	if err := deleteOrganization(ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	timeline := Timeline{
		UnderCompanyID: user.ActiveCompanyID,
		UserID:         user.ID,
		OrganizationID: ID,
		Action:         "deleted",
	}
	timeline.Name = model.Name
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal("ok")))
}

func handleGetOrganizationFields(w http.ResponseWriter, r *http.Request, user *User) {
}

func handlePostOrganizationFields(w http.ResponseWriter, r *http.Request, user *User) {
}

func handlePutOrganizationField(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleDeleteOrganizationField(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleGetOrganizationRelationships(w http.ResponseWriter, r *http.Request, user *User) {
}

func handlePostOrganizationRelationships(w http.ResponseWriter, r *http.Request, user *User) {
}

func handlePutOrganizationRelationship(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleDeleteOrganizationRelationship(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleGetContacts(w http.ResponseWriter, r *http.Request, user *User) {
	personID := r.URL.Query().Get("person_id")

	models, err := selectContactsByPerson(personID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Write(must(json.Marshal(models)))
}

func handlePostContacts(w http.ResponseWriter, r *http.Request, user *User) {
	var input Contact
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	input.DetectType()

	if err := insertContact(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	timeline := Timeline{
		UnderCompanyID: user.ActiveCompanyID,
		UserID:         user.ID,
		ContactID:      input.ID,
		Action:         "created",
	}
	timeline.Name = input.Name
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal(input)))
}

func handlePutContact(w http.ResponseWriter, r *http.Request, user *User) {
	var input Contact
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	input.DetectType()

	if err := updateContact(input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	timeline := Timeline{
		UnderCompanyID: user.ActiveCompanyID,
		UserID:         user.ID,
		ContactID:      input.ID,
		Action:         "updated",
	}
	timeline.Name = input.Name
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal(input)))
}

func handleDeleteContact(w http.ResponseWriter, r *http.Request, user *User) {
	vars := mux.Vars(r)
	ID := vars["id"]

	model, err := selectContactByID(ID)
	if err != nil {
		log.Println(err)
		http.Error(w, "error loading contact", http.StatusInternalServerError)
		return
	}

	if err := deleteContact(ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	timeline := Timeline{
		UnderCompanyID: user.ActiveCompanyID,
		UserID:         user.ID,
		ContactID:      model.ID,
		Action:         "deleted",
	}
	timeline.Name = model.Name
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal("ok")))
}

func handleGetPerson(w http.ResponseWriter, r *http.Request, user *User) {
	vars := mux.Vars(r)
	ID := vars["id"]

	model, err := selectPersonByID(ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.Write(must(json.Marshal(model)))
}

func handleGetPersons(w http.ResponseWriter, r *http.Request, user *User) {
	orgID := r.URL.Query().Get("org_id")

	var models []Person
	var err error

	if orgID != "" {
		models, err = selectPersonsByOrganization(orgID)
	} else {
		models, err = selectPersonsByCompany(user.ActiveCompanyID)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Write(must(json.Marshal(models)))
}

func handlePostPersons(w http.ResponseWriter, r *http.Request, user *User) {
	var input Person
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	input.CompanyID = user.ActiveCompanyID
	input.OwnerID = user.ID

	if err := assignOrganization(&input, *user); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := insertPerson(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	timeline := Timeline{
		UnderCompanyID: user.ActiveCompanyID,
		UserID:         user.ID,
		PersonID:       input.ID,
		Action:         "created",
	}
	timeline.Name = input.Name
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if input.Phone != "" {
		contact := Contact{
			PersonID: input.ID,
		}
		contact.Name = input.Phone

		timeline := Timeline{
			UnderCompanyID: user.ActiveCompanyID,
			UserID:         user.ID,
			ContactID:      contact.ID,
			Action:         "created",
		}
		timeline.Name = contact.Name
		if err := insertTimeline(&timeline); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if input.Email != "" {
		contact := Contact{
			PersonID: input.ID,
		}
		contact.Name = input.Email

		timeline := Timeline{
			UnderCompanyID: user.ActiveCompanyID,
			UserID:         user.ID,
			ContactID:      contact.ID,
			Action:         "created",
		}
		timeline.Name = contact.Name
		if err := insertTimeline(&timeline); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	model, err := selectPersonByID(input.ID)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal(model)))
}

func handlePutPerson(w http.ResponseWriter, r *http.Request, user *User) {
	var input Person
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := assignOrganization(&input, *user); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := updatePerson(input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	timeline := Timeline{
		UnderCompanyID: user.ActiveCompanyID,
		UserID:         user.ID,
		PersonID:       input.ID,
		Action:         "updated",
	}
	timeline.Name = input.Name
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	model, err := selectPersonByID(input.ID)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal(model)))
}

func handleDeletePerson(w http.ResponseWriter, r *http.Request, user *User) {
	vars := mux.Vars(r)
	ID := vars["id"]

	model, err := selectPersonByID(ID)
	if err != nil {
		log.Println(err)
		http.Error(w, "error loading person", http.StatusInternalServerError)
		return
	}

	if err := deletePerson(ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	timeline := Timeline{
		UnderCompanyID: user.ActiveCompanyID,
		UserID:         user.ID,
		PersonID:       ID,
		Action:         "deleted",
	}
	timeline.Name = model.Name
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal("ok")))
}

func handleGetPersonFields(w http.ResponseWriter, r *http.Request, user *User) {
}

func handlePostPersonFields(w http.ResponseWriter, r *http.Request, user *User) {
}

func handlePutPersonField(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleDeletePersonField(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleGetWorkflows(w http.ResponseWriter, r *http.Request, user *User) {
	models, err := selectWorkflowsByCompany(user.ActiveCompanyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.Write(must(json.Marshal(models)))
}

func handlePostWorkflows(w http.ResponseWriter, r *http.Request, user *User) {
	var input Workflow
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	input.CompanyID = user.ActiveCompanyID

	if err := insertWorkflow(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	timeline := Timeline{
		UnderCompanyID: user.ActiveCompanyID,
		UserID:         user.ID,
		WorkflowID:     input.ID,
		Action:         "created",
	}
	timeline.Name = input.Name
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal(input)))
}

func handlePutWorkflow(w http.ResponseWriter, r *http.Request, user *User) {
	vars := mux.Vars(r)
	ID := vars["id"]

	var input Workflow
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	model, err := selectWorkflowByID(ID)
	if err != nil {
		http.Error(w, "Error loading workflow", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	model.Name = input.Name

	if err := updateWorkflow(*model); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	timeline := Timeline{
		UnderCompanyID: model.CompanyID,
		UserID:         user.ID,
		WorkflowID:     model.ID,
		Action:         "updated",
	}
	timeline.Name = model.Name
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal(model)))
}

func handleDeleteWorkflow(w http.ResponseWriter, r *http.Request, user *User) {
	vars := mux.Vars(r)
	ID := vars["id"]

	workflow, err := selectWorkflowByID(ID)
	if err != nil {
		log.Println(err)
		http.Error(w, "error loading workflow", http.StatusInternalServerError)
		return
	}

	if err := deleteWorkflow(ID); err != nil {
		log.Println(err)
		http.Error(w, "error deleting workflow", http.StatusInternalServerError)
		return
	}

	timeline := Timeline{
		UnderCompanyID: workflow.CompanyID,
		UserID:         user.ID,
		WorkflowID:     ID,
		Action:         "deleted",
	}
	timeline.Name = workflow.Name
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal("ok")))
}

func handleGetPrices(w http.ResponseWriter, r *http.Request, user *User) {
}

func handlePostPrices(w http.ResponseWriter, r *http.Request, user *User) {
}

func handlePutPrice(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleDeletePrice(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleGetProducts(w http.ResponseWriter, r *http.Request, user *User) {
}

func handlePostProducts(w http.ResponseWriter, r *http.Request, user *User) {
}

func handlePutProduct(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleDeleteProduct(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleGetProductFields(w http.ResponseWriter, r *http.Request, user *User) {
}

func handlePostProductFields(w http.ResponseWriter, r *http.Request, user *User) {
}

func handlePutProductField(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleDeleteProductField(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleGetPushNotifications(w http.ResponseWriter, r *http.Request, user *User) {
}

func handlePostPushNotifications(w http.ResponseWriter, r *http.Request, user *User) {
}

func handlePutPushNotification(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleDeletePushNotification(w http.ResponseWriter, r *http.Request, user *User) {
}

func handleGetStages(w http.ResponseWriter, r *http.Request, user *User) {
	workflowID := r.URL.Query().Get("workflow_id")

	if workflowID == "" {
		workflowID = user.ActiveWorkflowID
	}

	models, err := selectStagesByWorkflow(workflowID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.Write(must(json.Marshal(models)))
}

func handlePostStages(w http.ResponseWriter, r *http.Request, user *User) {
	var input Stage
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := insertStage(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	timeline := Timeline{
		UnderCompanyID: user.ActiveCompanyID,
		UserID:         user.ID,
		StageID:        input.ID,
		Action:         "created",
	}
	timeline.Name = input.Name
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal(input)))
}

func handlePutStage(w http.ResponseWriter, r *http.Request, user *User) {
	vars := mux.Vars(r)
	ID := vars["id"]

	var input Stage
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	stage, err := selectStageByID(ID)
	if err != nil {
		http.Error(w, "Error loading stage", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	stage.OrderNr = input.OrderNr
	stage.Name = input.Name

	if err := updateStage(*stage); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	workflow, err := selectWorkflowByID(stage.WorkflowID)
	if err != nil {
		http.Error(w, "Error loading workflow", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	timeline := Timeline{
		UnderCompanyID: workflow.CompanyID,
		UserID:         user.ID,
		StageID:        stage.ID,
		Action:         "updated",
	}
	timeline.Name = stage.Name
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal(stage)))
}

func handleDeleteStage(w http.ResponseWriter, r *http.Request, user *User) {
	vars := mux.Vars(r)
	ID := vars["id"]

	stage, err := selectStageByID(ID)
	if err != nil {
		http.Error(w, "error loading stage", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	workflow, err := selectWorkflowByID(stage.WorkflowID)
	if err != nil {
		http.Error(w, "error loading workflow", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	if err := deleteStage(ID); err != nil {
		http.Error(w, "error deleting stage", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	timeline := Timeline{
		UnderCompanyID: workflow.CompanyID,
		UserID:         user.ID,
		StageID:        ID,
		Action:         "deleted",
	}
	timeline.Name = stage.Name
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal("ok")))
}

func handleGetTimeEntry(w http.ResponseWriter, r *http.Request, user *User) {
	vars := mux.Vars(r)
	ID := vars["id"]

	model, err := selectTimeEntryByID(ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Write(must(json.Marshal(model)))
}

func handleGetTimeEntries(w http.ResponseWriter, r *http.Request, user *User) {
	fromTime, err := parseTime(r.URL.Query().Get("time_entries_from"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println(err)
		return
	}

	untilTime, err := parseTime(r.URL.Query().Get("time_entries_until"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println(err)
		return
	}

	models, err := selectTimeEntriesByUserAndCompany(user.ID, user.ActiveCompanyID, fromTime, untilTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	format := r.URL.Query().Get("format")

	if "csv" != format {
		w.Write(must(json.Marshal(models)))
		return
	}

	exportCSV(w, models)
}

func handlePostTimeEntries(w http.ResponseWriter, r *http.Request, user *User) {
	var input TimeEntry
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	input.UserID = user.ID
	input.CompanyID = user.ActiveCompanyID

	if err := insertTimeEntry(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	timeline := Timeline{
		UnderCompanyID: input.CompanyID,
		UserID:         input.UserID,
		TimeEntryID:    input.ID,
		Action:         "created",
	}
	timeline.Name = input.Name
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal(input)))
}

func handlePutTimeEntry(w http.ResponseWriter, r *http.Request, user *User) {
	var input TimeEntry
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	input.UserID = user.ID
	input.CompanyID = user.ActiveCompanyID

	if err := updateTimeEntry(input); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	timeline := Timeline{
		UnderCompanyID: input.CompanyID,
		UserID:         user.ID,
		TimeEntryID:    input.ID,
		Action:         "updated",
	}
	timeline.Name = input.Name
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal(input)))
}

func handleDeleteTimeEntry(w http.ResponseWriter, r *http.Request, user *User) {
	vars := mux.Vars(r)
	ID := vars["id"]

	model, err := selectTimeEntryByID(ID)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := deleteTimeEntry(ID); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	timeline := Timeline{
		UnderCompanyID: model.CompanyID,
		UserID:         user.ID,
		TimeEntryID:    ID,
		Action:         "deleted",
	}
	timeline.Name = model.Name
	if err := insertTimeline(&timeline); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal("ok")))
}

func handleGetTokenLogin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["api_token"]

	if token == "" {
		return
	}

	ID, err := selectUserIDByApiToken(token)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session, _ := store.Get(r, sessionName)
	session.Values["user_id"] = ID
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound)
}

func handleGetUserEvents(w http.ResponseWriter, r *http.Request, user *User) {
	models, err := selectUserEventsByUser(user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Write(must(json.Marshal(models)))
}

func handlePostUserEvents(w http.ResponseWriter, r *http.Request, user *User) {
	var input UserEvent
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	input.UserID = user.ID

	if err := insertUserEvent(&input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(must(json.Marshal(input)))
}
