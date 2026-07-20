package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

const contentType = "application/json"

const devBaseURL = "http://localhost:8000"

const pingRoute = "/ping"
const pongRoute = "/pong"

const devPingURL = devBaseURL + pingRoute
const devPingsURL = devBaseURL + "/pings"
const devNewPingURL = devPingURL + "/new"

const devPongURL = devBaseURL + pongRoute
const devPongsURL = devBaseURL + "/pongs"

type NewPingRequest struct {
	Number uint `json:"number"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("No arguments provided.")
		printUsage()
		return
	}

	subcommand := os.Args[1]
	fmt.Println("Got:", subcommand)

	switch subcommand {
	case "ping":
		runGetPing(os.Args[2:])
	case "pings":
		runPings()
	case "pongs":
		runPongs()
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

func httpGetFrom(url string) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("HTTP GET request failed:", err)
		return
	}

	printResponse(resp)

	// Actually pretty print the response later
}

func httpPostTo(url, mediaType string, body io.Reader) {
	resp, err := http.Post(url, mediaType, body)
	if err != nil {
		fmt.Println("HTTP POST request failed:", err)
		return
	}

	printResponse(resp)
}

func printResponse(resp *http.Response) {
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Unexpected response:", resp.Status)
	}

	_, err := io.Copy(os.Stdout, resp.Body)
	if err != nil {
		fmt.Println("Could not read response:", err)
		return
	}

	fmt.Println()
}

func printUsage() {
	fmt.Println("Pretend this is a usage message.")
}

/*
To do:
- Send a create ping request and print the reply
- Implement subcommand syntax
- List pings X
- List pongs X
- Implement login
- Move password input to ReadPassword
- Implement token caching (in a config file, possibly hidden?)
- Implement logout
- Implement authentication status report
- Implement create pong request
- Implement change ping
- Implement change pong
- Implement delete ping
- Implement delete pong
- Implement usage message
*/
