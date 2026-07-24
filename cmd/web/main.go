package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

const appName = "Go Joe web client"

const webServerPort = "8080"

const webServerAddressFormat = ":%s"

func showHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, appName)
}

func main() {
	http.HandleFunc("/", showHome)

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
