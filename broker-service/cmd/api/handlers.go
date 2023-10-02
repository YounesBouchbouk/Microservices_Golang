package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/rpc"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   mailPayload `json:"mail,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type mailPayload struct {
	From        string   `json:"from"`
	To          string   `json:"to"`
	Subject     string   `json:"subject"`
	Message     string   `json:"message"`
	FromName    string   `json:"fromname"`
	Attachments []string `json:"attachments"`
}

func (app *Config) Brocker(w http.ResponseWriter, r *http.Request) {

	patload := jsonResponse{
		Error:   false,
		Message: "Hit the Brocker",
	}

	out, _ := json.MarshalIndent(patload, "", "\t")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write(out)

}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth)
	case "Log":
		app.logItemViaRPC(w, requestPayload.Log)
	case "mail":
		app.SendMail(w, requestPayload.Mail)
	default:
		app.errorJSON(w, errors.New("invalid methode"), http.StatusBadRequest)

	}

}

func (app *Config) SendMail(w http.ResponseWriter, mail mailPayload) {
	jsonData, _ := json.MarshalIndent(mail, "", "\t")

	request, err := http.NewRequest("POST", "http://mail-service/send", bytes.NewBuffer(jsonData))

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	client := http.Client{}

	response, err := client.Do(request)

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	defer response.Body.Close()

	// make shure we get the correct status

	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error calling service "))
		return
	}

	//create a variable we'll read response.Body into

	var jsonFromService jsonResponse

	//decode the json from the auth service

	err = json.NewDecoder(response.Body).Decode(&jsonFromService)

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		app.errorJSON(w, err, http.StatusUnauthorized)
		return

	}

	var payload jsonResponse

	payload.Error = false
	payload.Message = "email has been sent"
	payload.Data = jsonFromService.Data

	app.writeJSON(w, http.StatusAccepted, payload)

}

func (app *Config) logItem(w http.ResponseWriter, entry LogPayload) {
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	//call the service

	request, err := http.NewRequest("POST", "http://logger-service/log", bytes.NewBuffer(jsonData))

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	defer response.Body.Close()

	// make shure we get the correct status

	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error calling service "))
		return
	}

	//create a variable we'll read response.Body into

	var payload jsonResponse

	payload.Error = false
	payload.Message = "Logged!"

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
	jsonData, _ := json.MarshalIndent(a, "", "\t")

	//call the service

	request, err := http.NewRequest("POST", "http://authentiation-service/authenticate", bytes.NewBuffer(jsonData))

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	client := http.Client{}

	response, err := client.Do(request)

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	defer response.Body.Close()

	// make shure we get the correct status

	if response.StatusCode == http.StatusUnauthorized {
		app.errorJSON(w, errors.New("invalid credentials"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error calling service "))
		return
	}

	//create a variable we'll read response.Body into

	var jsonFromService jsonResponse

	//decode the json from the auth service

	err = json.NewDecoder(response.Body).Decode(&jsonFromService)

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		app.errorJSON(w, err, http.StatusUnauthorized)
		return

	}

	var payload jsonResponse

	payload.Error = false
	payload.Message = "Authenticated!"
	payload.Data = jsonFromService.Data

	app.writeJSON(w, http.StatusAccepted, payload)
}

type RPCPayload struct {
	Name string
	Data string
}

func (app *Config) logItemViaRPC(w http.ResponseWriter, entry LogPayload) {

	client, err := rpc.Dial("tcp", "logger-service:5001")

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	payloadRpc := RPCPayload{
		Name: entry.Name,
		Data: entry.Data,
	}

	var result string

	client.Call("RPCServer.LogInfo", payloadRpc, &result)

	if err != nil {
		app.errorJSON(w, err)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: result,
	}

	app.writeJSON(w, http.StatusAccepted, payload)

}
