package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	serv "github.com/subtlepseudonym/go-utils/http"

	"squawker/service"
	"squawker/util"
)

var currentId string

func checkForAndPlayNext() {
	nextCalled := service.GetPlayNextChan()
	for {
		// Mostly I think this if condition is cumbersome, but it uses well-defined string equality
		if util.GetNowPlaying() == nil || util.GetNowPlaying().Id == currentId {
			// When checkForAndPlayNext() is first called, this blocks (because util.GetMedia() and currentMedia are both nil)
			util.PlayNext()         // This is where the magic happens
			time.Sleep(time.Second) // player must start playing, or util.GetMediaLength() returns 0
		}
		currentId = util.GetNowPlaying().Id

		// This multiplies the total media length by the percentage played
		timeToWaitMs := int(float32(util.GetMediaLength()) * (1.0 - util.GetMediaPosition()))

		// This keeps async calls to PlayNext() from leaving us with an awkward pause after songs
		select {
		case <-nextCalled:
			time.Sleep(time.Second)
			continue
		case <-time.After(time.Duration(timeToWaitMs) * time.Millisecond):
			continue
		}
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
