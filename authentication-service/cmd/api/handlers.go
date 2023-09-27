package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func (app *Config) authenticate(w http.ResponseWriter, r *http.Request) {
	var resuestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &resuestPayload)

	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	//validate the user against the db

	user, err := app.Models.User.GetByEmail(resuestPayload.Email)

	if err != nil {
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)

		return
	}

	valid, err := user.PasswordMatches(resuestPayload.Password)

	if err != nil || !valid {
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	// log authentication

	err = app.logRequest("authentication", fmt.Sprintf("%s logged in", user.Email))

	if err != nil {
		app.errorJSON(w, errors.New("can't log data "), http.StatusBadRequest)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    user,
	}

	app.writeJSON(w, http.StatusAccepted, payload)

}

func (app *Config) logRequest(name, data string) error {
	var entry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	entry.Name = name
	entry.Data = data

	jsonData, _ := json.MarshalIndent(entry, "", "\t")
	logServicesURL := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServicesURL, bytes.NewBuffer(jsonData))

	if err != nil {
		return err
	}

	client := &http.Client{}

	client.Do(request)

	if err != nil {
		return err
	}

	return nil

}
