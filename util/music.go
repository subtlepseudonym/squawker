package util

import (
	"log"

	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
)

const fadeTime int = 1000 // in milliseconds
const mixFlags int = mix.INIT_OGG | mix.INIT_MP3
const defaultFrequency int = 48000
const defaultAudioFormat uint16 = sdl.AUDIO_F32SYS
const defaultChunkSize int = 2048 // higher numbers are fine for music

// PlayNext() blocks, so we don't want multiple instances running at once
var playNextAlreadyQueued bool = false

// This function assumes that the next QueuedAudioInfo has already been downloaded
func PlayNext() {
	if playNextAlreadyQueued {
		return
	} else {
		playNextAlreadyQueued = true
	}

	queuedAudioFileInfo := DequeueAudioInfo()
	lastAudioFileInfo := GetNowPlaying()
	if lastAudioFileInfo != nil {
		lastAudioFileInfo.Mus.Free()
		addToLog(*lastAudioFileInfo)
	}

	nowPlaying = queuedAudioFileInfo
	mus, err := mix.LoadMUS(nowPlaying.Filename)
	if err != nil {
		log.Printf("Error loading music file %s\n", nowPlaying.Filename)
	}
	nowPlaying.Mus = mus
	mus.FadeIn(1, fadeTime)

	log.Printf("Now playing: %s\n", nowPlaying.Title)
	log.Println("Log --", GetPrettyAudioLogString(5))

	playNextAlreadyQueued = false
}

func init() {
	// This stuff is pretty sluggish
	err := sdl.Init(sdl.INIT_AUDIO)
	if err != nil {
		panic(err)
	}

	flags := mixFlags
	err = mix.Init(flags)
	if err != nil {
		panic(err)
	}

	// Default values are 22050, AUDIO_S16SYS, 2, 1024
	err = mix.OpenAudio(defaultFrequency, defaultAudioFormat, mix.DEFAULT_CHANNELS, defaultChunkSize)
	if err != nil {
		panic(err)
	}
}
