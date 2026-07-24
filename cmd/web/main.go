package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"time"
)

const appName = "Go Joe web client"

const devBaseURL = "http://localhost:8000"
const devPingsURL = devBaseURL + "/pings"
const webServerPort = "8080"

const webServerAddressFormat = ":%s"

var homeTemplate = template.Must(
	template.ParseFiles("cmd/web/templates/home.html"),
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

func main() {
	http.HandleFunc("/", showHome)
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

	err = pingsTemplate.Execute(w, records)
	if err != nil {
		log.Println("Could not render pings template:", err)
	}
}
