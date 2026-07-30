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
const tokenCookieName = "token"
const webServerPort = "8080"
const webServerAddressFormat = ":%s"

const devBaseURL = "http://localhost:8000"

const devAuthURL = devBaseURL + "/auth"
const devLoginURL = devBaseURL + "/login"
const devLogoutURL = devBaseURL + "/logout"

const devNewPingURL = devPingURL + "/new"
const devPingURL = devBaseURL + "/ping"
const devPingsURL = devBaseURL + "/pings"

const devNewPongURL = devPongURL + "/new"
const devPongURL = devBaseURL + "/pong"
const devPongsURL = devBaseURL + "/pongs"

var homeTemplate = template.Must(
	template.ParseFiles("cmd/web/templates/home.html"),
)

type AuthResponse struct {
	Status string `json:"status"`
}

var loginTemplate = template.Must(
	template.ParseFiles("cmd/web/templates/login.html"),
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

var newPongTemplate = template.Must(
	template.ParseFiles("cmd/web/templates/new_pong.html"),
)

var pongTemplate = template.Must(
	template.ParseFiles("cmd/web/templates/pong.html"),
)

var pongsTemplate = template.Must(
	template.ParseFiles("cmd/web/templates/pongs.html"),
)

type PingRecord struct {
	ID        uint       `json:"ID"`
	CreatedAt time.Time  `json:"CreatedAt"`
	UpdatedAt time.Time  `json:"UpdatedAt"`
	DeletedAt *time.Time `json:"DeletedAt"`
	Time      time.Time  `json:"time"`
	Number    uint       `json:"number"`
}

type PongRecord struct {
	ID        uint       `json:"ID"`
	CreatedAt time.Time  `json:"CreatedAt"`
	UpdatedAt time.Time  `json:"UpdatedAt"`
	DeletedAt *time.Time `json:"DeletedAt"`
	Time      time.Time  `json:"time"`
	Number    uint       `json:"number"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type NewPingRequest struct {
	Number uint `json:"number"`
}

type NewPongRequest struct {
	Number uint `json:"number"`
}

func main() {
	http.HandleFunc("/", showHome)

	http.HandleFunc("GET /auth", showAuth)
	http.HandleFunc("GET /login", showLogin)
	http.HandleFunc("POST /login", login)
	http.HandleFunc("POST /logout", logout)

	http.HandleFunc("POST /ping/{id}/delete", deletePing)
	http.HandleFunc("GET /ping/{id}", showPing)
	http.HandleFunc("GET /ping/new", showNewPing)
	http.HandleFunc("POST /ping/new", createPing)
	http.HandleFunc("/pings", showPings)

	http.HandleFunc("GET /pong/{id}", showPong)
	http.HandleFunc("GET /pong/new", showNewPong)
	http.HandleFunc("POST /pong/new", createPong)
	http.HandleFunc("GET /pongs", showPongs)

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

func login(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		http.Error(
			w,
			"Username and password are required.",
			http.StatusBadRequest,
		)
		return
	}

	request := LoginRequest{Username: username, Password: password}
	body, err := json.Marshal(request)
	if err != nil {
		http.Error(
			w,
			"Could not encode login request.",
			http.StatusInternalServerError,
		)
		return
	}

	resp, err := http.Post(devLoginURL, contentType, bytes.NewReader(body))
	if err != nil {
		http.Error(w, "Could not log in.", http.StatusBadGateway)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		http.Error(w, "Invalid username or password.", http.StatusUnauthorized)
		return
	}

	if resp.StatusCode != http.StatusOK {
		message := "Login API returned " + resp.Status
		http.Error(w, message, http.StatusBadGateway)
		return
	}

	response := LoginResponse{}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&response)
	if err != nil {
		http.Error(
			w,
			"Could not decode login response.",
			http.StatusBadGateway,
		)
		return
	}

	if response.Token == "" {
		http.Error(
			w,
			"Login response did not contain a token.",
			http.StatusBadGateway,
		)
		return
	}

	tokenCookie := http.Cookie{
		Name:     tokenCookieName,
		Value:    response.Token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &tokenCookie)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func logout(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(devLogoutURL)
	if err != nil {
		http.Error(w, "Could not log out.", http.StatusBadGateway)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		message := "Logout API returned " + resp.Status
		http.Error(w, message, http.StatusBadGateway)
		return
	}

	tokenCookie := http.Cookie{
		Name:     tokenCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	}
	http.SetCookie(w, &tokenCookie)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func showHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err := homeTemplate.Execute(w, appName)
	if err != nil {
		log.Println("Could not render home template:", err)
	}
}

func showAuth(w http.ResponseWriter, r *http.Request) {
	tokenCookie, err := r.Cookie(tokenCookieName)
	if err != nil {
		http.Error(w, "Not logged in.", http.StatusUnauthorized)
		return
	}

	request, err := http.NewRequest(http.MethodGet, devAuthURL, nil)
	if err != nil {
		http.Error(
			w,
			"Could not create authentication request.",
			http.StatusInternalServerError,
		)
		return
	}

	request.Header.Set("Authorization", tokenCookie.Value)

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		http.Error(w, "Could not check authentication.", http.StatusBadGateway)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		http.Error(w, "Not logged in.", http.StatusUnauthorized)
		return
	}

	if resp.StatusCode != http.StatusOK {
		message := "Auth API returned " + resp.Status
		http.Error(w, message, http.StatusBadGateway)
		return
	}

	response := AuthResponse{}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&response)
	if err != nil {
		http.Error(
			w,
			"Could not decode authentication response.",
			http.StatusBadGateway,
		)
		return
	}

	if response.Status == "" {
		http.Error(
			w,
			"Authentication response did not contain a status.",
			http.StatusBadGateway,
		)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, "Authenticated as", response.Status)
}

func showLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err := loginTemplate.Execute(w, appName)
	if err != nil {
		log.Println("Could not render login template:", err)
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

func deletePing(w http.ResponseWriter, r *http.Request) {
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

	deletePingURL := devPingURL + "/" + strconv.Itoa(id) + "/delete"

	resp, err := http.Post(deletePingURL, "", nil)
	if err != nil {
		http.Error(w, "Could not delete ping.", http.StatusBadGateway)
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

func showNewPong(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	err := newPongTemplate.Execute(w, appName)
	if err != nil {
		log.Println("Could not render new pong template:", err)
	}
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
		http.Error(w, "Could not decode ping response.", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	data := map[string]any{"AppName": appName, "Record": record}
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

	data := map[string]any{"AppName": appName, "Records": records}
	err = pingsTemplate.Execute(w, data)
	if err != nil {
		log.Println("Could not render pings template:", err)
	}
}

func createPong(w http.ResponseWriter, r *http.Request) {
	numberValue := r.FormValue("number")
	number, err := strconv.Atoi(numberValue)
	if err != nil || number < 1 {
		http.Error(
			w,
			"Pong number value must be a positive integer.",
			http.StatusBadRequest,
		)
		return
	}

	request := NewPongRequest{Number: uint(number)}
	body, err := json.Marshal(request)
	if err != nil {
		http.Error(
			w,
			"Could not encode pong request.",
			http.StatusInternalServerError,
		)
		return
	}

	tokenCookie, err := r.Cookie(tokenCookieName)
	if err != nil {
		http.Error(w, "Not logged in.", http.StatusUnauthorized)
		return
	}

	httpRequest, err := http.NewRequest(
		http.MethodPost,
		devNewPongURL,
		bytes.NewReader(body),
	)
	if err != nil {
		http.Error(
			w,
			"Could not create pong request.",
			http.StatusInternalServerError,
		)
		return
	}

	httpRequest.Header.Set("Content-Type", contentType)
	httpRequest.Header.Set("Authorization", tokenCookie.Value)

	resp, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		http.Error(w, "Could not create pong.", http.StatusBadGateway)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		http.Error(w, "Not logged in.", http.StatusUnauthorized)
		return
	}

	if resp.StatusCode != http.StatusOK {
		message := "Pong API returned " + resp.Status
		http.Error(w, message, http.StatusBadGateway)
		return
	}

	http.Redirect(w, r, "/pongs", http.StatusSeeOther)
}

func showPong(w http.ResponseWriter, r *http.Request) {
	idValue := r.PathValue("id")

	id, err := strconv.Atoi(idValue)
	if err != nil || id < 1 {
		http.Error(
			w,
			"Pong ID must be a positive integer.",
			http.StatusBadRequest,
		)
		return
	}

	pongURL := devPongURL + "/" + strconv.Itoa(id)
	resp, err := http.Get(pongURL)
	if err != nil {
		http.Error(w, "Could not retrieve pong.", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		message := "Pong API returned " + resp.Status
		http.Error(w, message, http.StatusBadGateway)
		return
	}

	record := PongRecord{}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&record)
	if err != nil {
		http.Error(w, "Could not decode pong response.", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	data := map[string]any{"AppName": appName, "Record": record}
	err = pongTemplate.Execute(w, data)
	if err != nil {
		log.Println("Could not render pong template:", err)
	}
}

func showPongs(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(devPongsURL)
	if err != nil {
		http.Error(w, "Could not retrieve pongs.", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		message := "Pongs API returned " + resp.Status
		http.Error(w, message, http.StatusBadGateway)
		return
	}

	records := []PongRecord{}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&records)
	if err != nil {
		http.Error(
			w,
			"Could not decode pongs response.",
			http.StatusBadGateway,
		)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	data := map[string]any{"AppName": appName, "Records": records}
	err = pongsTemplate.Execute(w, data)
	if err != nil {
		log.Println("Could not render pongs template:", err)
	}
}
