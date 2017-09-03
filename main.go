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
var nextCalled chan bool = service.GetPlayNextChan()
var toggleCalled chan bool = service.GetTogglePlaybackChan()

// This is less of a good example of how to control playback properly and more an exercise in getting better with channels and timing
func checkForAndPlayNext() {
	for {
		nowPlaying := util.GetNowPlaying()
		// Mostly I think this if condition is cumbersome, but it uses well-defined string equality
		if nowPlaying == nil || nowPlaying.Id == currentId {
			// When checkForAndPlayNext() is first called, this blocks (because util.GetMedia() and currentMedia are both nil)
			util.PlayNext()         // This is where the magic happens
			time.Sleep(time.Second) // player must start playing, or util.GetMediaLength() returns 0
		}
		currentId = util.GetNowPlaying().Id
		getTimeRemainingAndWait()
	}
}

func getTimeRemainingAndWait() {
	// This multiplies the total media length by the percentage played
	timeToWaitMs := int(util.GetNowPlaying().Duration/time.Millisecond) - int(float32(util.GetMediaLength())*util.GetMediaPosition())

	// This keeps async calls to PlayNext() from leaving us with an awkward pause after songs
	select {
	case <-nextCalled:
		time.Sleep(time.Second)
	case <-time.After(time.Duration(timeToWaitMs) * time.Millisecond):
	case <-toggleCalled:
		if !util.PlayerIsPlaying() {
			<-toggleCalled // just wait until toggle is called again
		}
		getTimeRemainingAndWait()
	}
}

func main() {
	defer util.CleanUp() // TODO: write a hook to run this call on ^C
	go checkForAndPlayNext()

	http.HandleFunc("/add", service.AddHandler)
	http.HandleFunc("/next", service.NextHandler)
	http.HandleFunc("/toggle", service.TogglePlaybackHandler)

	serv.LogPublicIpAddress(nil, util.GetPort())
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", util.GetPort()), nil))
}
