package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"models"
	"net/http"
)

// UsersHandler returns a list of the users
func UsersHandler(w http.ResponseWriter, r *http.Request) {
	apikey := r.URL.Query().Get("apikey")
	if apikey != Constants.APIKey {
		log.Printf("invalid apikey: {%v}", apikey)
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("invalid apikey"))
		return
	}

	users, err := GetAllUsers()
	if err != nil {
		log.Printf("problem accessing database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("problem accessing database"))
		return
	}

	b, err := json.Marshal(users)
	if err != nil {
		log.Printf("problem marshalling data: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("problem marshalling data"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

// EventHandler figures out what type of event was triggered and routes the event to the correct handler
func EventHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	event := &models.EventModel{}
	err = json.Unmarshal(body, event)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	if event.Token != Constants.Token {
		err = fmt.Errorf("Invalid token: {%v}", event.Token)
		log.Printf("Invalid token: %v", err)
		http.Error(w, "invalid token", http.StatusForbidden)
		return
	}

	if event.Type == "url_verification" {
		URLVerificationHandler(w, body)
		return
	} else if event.Event.Type == "user_change" || event.Event.Type == "team_join" {
		UserChangeHandler(w, body)
		return
	} else {
		log.Printf("Invalid event: %v", event.Event.Type)
		http.Error(w, "Invalid event", http.StatusBadRequest)
		return
	}
}

// UserChangeHandler responds to slack's url verification event
var UserChangeHandler = func(w http.ResponseWriter, body []byte) {
	userChangeRequest := &models.UserChangeRequest{}
	err := json.Unmarshal(body, userChangeRequest)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	err = InsertOrUpdateUser(userChangeRequest.Event.User)
	if err != nil {
		log.Printf("Error updating database: %v", err)
		http.Error(w, "error updating database", http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

// URLVerificationHandler responds to slack's url verification event
var URLVerificationHandler = func(w http.ResponseWriter, body []byte) {
	verification := &models.URLVerificationModel{}
	err := json.Unmarshal(body, verification)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	if verification.Token != Constants.Token {
		err = fmt.Errorf("Invalid token: {%v}", verification.Token)
		log.Printf("Invalid token: %v", err)
		http.Error(w, "invalid token", http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(verification.Challenge))
}
