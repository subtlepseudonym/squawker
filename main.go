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
			util.PlayNext()         // This is where the magic happens
			time.Sleep(time.Second) // player must start playing, or util.GetMediaLength() returns 0
		}
		currentMedia = util.GetMedia()
		// This multiplies the total media length by the percentage played
		timeToWaitMs := int(float32(util.GetMediaLength()) * (1.0 - util.GetMediaPosition()))
		time.Sleep(time.Duration(timeToWaitMs) * time.Millisecond)
	}
}

func main() {
	defer util.CleanUp() // TODO: write a hook to run this call on ^C
	go checkForAndPlayNext()

	http.HandleFunc("/add", service.AddHandler)
	http.HandleFunc("/next", service.NextHandler)

	serv.LogPublicIpAddress(nil, util.GetPort())
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", util.GetPort()), nil))
}
