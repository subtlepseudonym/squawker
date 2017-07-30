package main

import (
	"fmt"
	"log"
	"net/http"

	serv "github.com/subtlepseudonym/go-utils/http"

	"squawker/service"
	"squawker/util"
)

func main() {
	serv.LogPublicIpAddress(nil, util.GetPort())

	http.HandleFunc("/add", service.AddHandler)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", util.GetPort()), nil))
}
