package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"golang.org/x/term"
)

const contentType = "application/json"
const devBaseURL = "http://localhost:8000"
const tokenPath = ".go-joe-token"

const authRoute = "/auth"
const loginRoute = "/login"
const logoutRoute = "/logout"
const pingRoute = "/ping"
const pongRoute = "/pong"

const devAuthURL = devBaseURL + authRoute
const devLoginURL = devBaseURL + loginRoute
const devLogoutURL = devBaseURL + logoutRoute

const devPingURL = devBaseURL + pingRoute
const devPingsURL = devBaseURL + "/pings"
const devNewPingURL = devPingURL + "/new"

const devPongURL = devBaseURL + pongRoute
const devPongsURL = devBaseURL + "/pongs"

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

type readCloser struct {
	io.Reader
	io.Closer
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("No arguments provided.")
		printUsage()
		return
	}

	subcommand := os.Args[1]
	switch subcommand {
	case "auth":
		runAuth()
	case "login":
		runLogin(os.Args[2:])
	case "logout":
		runLogout()
	case "ping":
		runPing(os.Args[2:])
	case "pings":
		runPings()
	case "pongs":
		runPongs()
	default:
		fmt.Printf("%q not recognized.\n", subcommand)
		printUsage()
	}
}

func runAuth() {
	token, err := os.ReadFile(tokenPath) // Factor this into a helper
	if err != nil {
		fmt.Println("Could not read saved token:", err)
		return
	}

	// This too
	request, err := http.NewRequest(http.MethodGet, devAuthURL, nil)
	if err != nil {
		fmt.Println("Could not create auth request:", err)
		return
	}

	request.Header.Set("Authorization", string(token))

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		fmt.Println("Auth request failed:", err)
		return
	}
	err = printResponse(resp)
    if err != nil {
        return
    }
}

func runLogin(args []string) {
	if len(args) < 1 {
		fmt.Println("login: username is required.")
		return
	}

	// Validate username string from args[0] here

	fmt.Print("Enter password: ")
	passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		fmt.Println("Couldn't read password.")
		return
	}

	request := LoginRequest{Username: args[0], Password: string(passwordBytes)}
	body, err := json.Marshal(request)
	if err != nil {
		fmt.Println("Unexpected failure marshalling request to JSON:", err)
		return
	}

	responseBytes, status, err := httpPostTo(
		devLoginURL,
		contentType,
		bytes.NewReader(body),
	)
	if err != nil {
		return
	}

	if status != http.StatusOK {
		return
	}

	response := LoginResponse{}
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		fmt.Println("Could not decode login response:", err)
		return
	}

	if response.Token == "" {
		fmt.Println("Login response did not contain a token.")
		return
	}

	saveTokenToFile([]byte(response.Token))
}

func runLogout() {
	_, status, err := httpGetFrom(devLogoutURL)
	if err != nil {
		return
	}

	if status != http.StatusOK {
		return
	}

	err = os.Remove(tokenPath)
	if errors.Is(err, os.ErrNotExist) {
	    return
	}
	if err != nil {
		fmt.Println("Could not remove saved token:", err)
		return
	}
}

func runPing(args []string) {
	if len(args) < 1 {
		fmt.Println("ping: subcommand requires argument.")
		return
	}

	subcommand := args[0]
	switch subcommand {
	case "get":
		runGetPing(args[1:])
	case "new":
		runNewPing(args[1:])
	default:
		fmt.Printf("%q not recognized.\n", subcommand)
		printUsage()
	}
}

func runGetPing(args []string) {
	if len(args) < 1 {
		fmt.Println("ping: subcommand requires argument.")
		return
	}

	id, err := strconv.Atoi(args[0])
	if err != nil || id < 1 {
		fmt.Println("ping: subcommand ID argument must be a positive integer.")
		return
	}

	getPingURL := devPingURL + "/" + strconv.Itoa(id)
	httpGetFrom(getPingURL)
}

func runNewPing(args []string) {
	if len(args) < 1 {
		fmt.Println("ping: subcommand requires argument.")
		return
	}

	num, err := strconv.Atoi(args[0])
	if err != nil || num < 1 {
		fmt.Println("ping: subcommand argument must be a positive integer.")
		return
	}

	request := NewPingRequest{Number: uint(num)}
	body, err := json.Marshal(request)
	if err != nil {
		fmt.Println("Unexpected failure marshalling request to JSON:", err)
		return
	}

	httpPostTo(devNewPingURL, contentType, bytes.NewReader(body))
}

func runPings() {
	httpGetFrom(devPingsURL)
}

func runPongs() {
	httpGetFrom(devPongsURL)
}

func httpGetFrom(url string) ([]byte, int, error) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("HTTP GET request failed:", err)
		return nil, 0, err
	}

	status := resp.StatusCode

	var responseBody bytes.Buffer
	bodyReader := io.TeeReader(resp.Body, &responseBody)
	resp.Body = readCloser{Reader: bodyReader, Closer: resp.Body}

	err = printResponse(resp)
	if err != nil {
		return nil, status, err
	}

	// Actually pretty print the response later

	return responseBody.Bytes(), status, nil
}

func httpPostTo(url, mediaType string, body io.Reader) ([]byte, int, error) {
	resp, err := http.Post(url, mediaType, body)
	if err != nil {
		fmt.Println("HTTP POST request failed:", err)
		return nil, 0, err
	}

	status := resp.StatusCode

	var responseBody bytes.Buffer
	bodyReader := io.TeeReader(resp.Body, &responseBody)
	resp.Body = readCloser{Reader: bodyReader, Closer: resp.Body}

	err = printResponse(resp)
	if err != nil {
		return nil, status, err
	}

	return responseBody.Bytes(), status, nil
}

func printResponse(resp *http.Response) error {
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Unexpected response:", resp.Status)
	}

	_, err := io.Copy(os.Stdout, resp.Body)
	if err != nil {
		fmt.Println("Could not read response:", err)
		return err
	}

	fmt.Println()
	return nil
}

func saveTokenToFile(value []byte) {
	err := os.WriteFile(tokenPath, value, 0600)
	if err != nil {
		fmt.Println("Could not write token config file:", err)
		return
	}
}

func printUsage() {
	fmt.Println("Pretend this is a usage message.")
}

/*
To do:
- Send a create ping request and print the reply X
- Implement subcommand syntax X
- List pings X
- List pongs X
- Implement login X
- Move password input to ReadPassword X
- Implement token caching (in a config file, possibly hidden?) X
- Implement logout X
- Implement authentication status report X
- Implement create pong request
- Implement change ping
- Implement change pong
- Implement delete ping
- Implement delete pong
- Implement usage message
*/
