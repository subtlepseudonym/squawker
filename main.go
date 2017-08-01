package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/adrg/libvlc-go"
	serv "github.com/subtlepseudonym/go-utils/http"

	"squawker/service"
	"squawker/util"
)

var currentMedia *vlc.Media

func checkForAndPlayNext() {
	for {
		if util.GetMedia() == currentMedia {
			// When checkForAndPlayNext() is first called, this blocks (because util.GetMedia() and currentMedia are both nil)
			util.PlayNext()
			time.Sleep(time.Second) // player must start playing, or util.GetMediaLength() returns 0
		}
		currentMedia = util.GetMedia()
		time.Sleep(time.Duration(util.GetMediaLength()) * time.Millisecond)
	}
}

func main() {
	defer util.CleanUp()
	go checkForAndPlayNext()

	http.HandleFunc("/add", service.AddHandler)
	http.HandleFunc("/next", service.NextHandler)

	serv.LogPublicIpAddress(nil, util.GetPort())
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", util.GetPort()), nil))
}
