package main

import (
	"log"
	"net/http"

	serv "github.com/subtlepseudonym/go-utils/http"
)

const defaultPort = ":15567"

var port string

func init() {
	port = defaultPort

	serv.LogPublicIpAddress(nil)
}

func main() {
	log.Fatal(http.ListenAndServe(port, nil))
}
