package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	serv "github.com/subtlepseudonym/go-utils/http"
	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"

	"squawker/service"
	"squawker/util"
)

const checkAudioStatusFrequency = 250 // every 250 ms (roughly)

func checkAudioStatus() {
	for {
		if !mix.PlayingMusic() && mix.FadingMusic() == mix.NO_FADING {
			util.PlayNext()
		}
		time.Sleep(checkAudioStatusFrequency * time.Millisecond)
	}
}

func main() {
	defer sdl.Quit()
	serv.LogPublicIpAddress(nil, util.GetPort())

	// Check audio status every
	go checkAudioStatus()

	http.HandleFunc("/add", service.AddHandler)
	http.HandleFunc("/next", service.NextHandler)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", util.GetPort()), nil))
}
