package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	serv "github.com/subtlepseudonym/go-utils/http"

	"squawker/service"
)

const defaultPort = 15567
const defaultFileBacklogSize = 5 // maintain the last 5 songs played, delete anything older
const defaultMaxSongLength = 600 // length in seconds

var port int
var fileBacklogSize int
var maxSongLength int

func init() {
	flag.IntVar(&port, "port", defaultPort, "Set the port where the server listens for incoming connections")
	flag.IntVar(&fileBacklogSize, "backlog-size", defaultFileBacklogSize, "Only this many files will be stored with oldest files being deleted first")
	flag.IntVar(&maxSongLength, "max-length", defaultMaxSongLength, "This sets the maximum song length allowed")

	// TODO: create directory where temp .mp3 files will be stored
	// TODO: will likely need a util package where flags are parsed and general book keeping is done
}

func main() {
	flag.Parse()
	serv.LogPublicIpAddress(nil, port)

	http.HandleFunc("/next", service.NextHandler)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
