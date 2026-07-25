package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

const appName = "Go Joe web client"
const contentType = "application/json"
const webServerPort = "8080"
const webServerAddressFormat = ":%s"

const devBaseURL = "http://localhost:8000"

const devNewPingURL = devPingURL + "/new"
const devPingURL = devBaseURL + "/ping"
const devPingsURL = devBaseURL + "/pings"

var homeTemplate = template.Must(
	template.ParseFiles("cmd/web/templates/home.html"),
)

var newPingTemplate = template.Must(
	template.ParseFiles("cmd/web/templates/new_ping.html"),
)

var pingTemplate = template.Must(
	template.ParseFiles("cmd/web/templates/ping.html"),
)

var pingsTemplate = template.Must(
	template.ParseFiles("cmd/web/templates/pings.html"),
)

type PingRecord struct {
	ID        uint       `json:"ID"`
	CreatedAt time.Time  `json:"CreatedAt"`
	UpdatedAt time.Time  `json:"UpdatedAt"`
	DeletedAt *time.Time `json:"DeletedAt"`
	Time      time.Time  `json:"time"`
	Number    uint       `json:"number"`
}

type NewPingRequest struct {
	Number uint `json:"number"`
}

func main() {
	http.HandleFunc("/", showHome)
	http.HandleFunc("GET /ping/{id}", showPing)
	http.HandleFunc("GET /ping/new", showNewPing)
	http.HandleFunc("POST /ping/new", createPing)
	http.HandleFunc("/pings", showPings)

	serverAddress := fmt.Sprintf(webServerAddressFormat, webServerPort)
	host, port, err := net.SplitHostPort(serverAddress)
	if err != nil {
		log.Fatal(err)
	}

	if host == "" {
		host = "localhost"
	}

	serverURL := "http://" + net.JoinHostPort(host, port)
	fmt.Printf("%s: starting server on %s\n", appName, serverURL)

	err = http.ListenAndServe(serverAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func showHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err := homeTemplate.Execute(w, appName)
	if err != nil {
		log.Println("Could not render home template:", err)
	}
}

func showNewPing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err := newPingTemplate.Execute(w, appName)
	if err != nil {
		log.Println("Could not render new ping template:", err)
	}
}

func createPing(w http.ResponseWriter, r *http.Request) {
	numberValue := r.FormValue("number")
	number, err := strconv.Atoi(numberValue)
	if err != nil || number < 1 {
		http.Error(
			w,
			"Ping number value must be a positive integer.",
			http.StatusBadRequest,
		)
		return
	}

	request := NewPingRequest{Number: uint(number)}
	body, err := json.Marshal(request)
	if err != nil {
		http.Error(
			w,
			"Could not encode ping request.",
			http.StatusInternalServerError,
		)
		return
	}

	resp, err := http.Post(devNewPingURL, contentType, bytes.NewReader(body))
	if err != nil {
		http.Error(w, "Could not create ping.", http.StatusBadGateway)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		message := "Ping API returned " + resp.Status
		http.Error(w, message, http.StatusBadGateway)
		return
	}

	http.Redirect(w, r, "/pings", http.StatusSeeOther)
}

func showPing(w http.ResponseWriter, r *http.Request) {
	idValue := r.PathValue("id")

	id, err := strconv.Atoi(idValue)
	if err != nil || id < 1 {
		http.Error(
			w,
			"Ping ID must be a positive integer.",
			http.StatusBadRequest,
		)
		return
	}

	pingURL := devPingURL + "/" + strconv.Itoa(id)
	resp, err := http.Get(pingURL)
	if err != nil {
		http.Error(w, "Could not retrieve ping.", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		message := "Ping API returned " + resp.Status
		http.Error(w, message, http.StatusBadGateway)
		return
	}

	record := PingRecord{}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&record)
	if err != nil {
		http.Error(
			w,
			"Could not decode ping response.",
			http.StatusBadGateway,
		)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	data := map[string]any{
		"AppName": appName,
		"Record":  record,
	}
	err = pingTemplate.Execute(w, data)
	if err != nil {
		log.Println("Could not render ping template:", err)
	}
}

func showPings(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(devPingsURL)
	if err != nil {
		http.Error(w, "Could not retrieve pings.", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		message := "Pings API returned " + resp.Status
		http.Error(w, message, http.StatusBadGateway)
		return
	}

	records := []PingRecord{}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&records)
	if err != nil {
		http.Error(
			w,
			"Could not decode pings response.",
			http.StatusBadGateway,
		)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	data := map[string]any{
		"AppName": appName,
		"Records": records,
	}
	err = pingsTemplate.Execute(w, data)
	if err != nil {
		log.Println("Could not render pings template:", err)
	}
}
