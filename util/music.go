package util

import (
	"log"

	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
)

const fadeTime int = 1000 // in milliseconds
const mixFlags int = mix.INIT_OGG | mix.INIT_MP3

// This function assumes that the next QueuedAudioInfo has already been downloaded
func PlayNext() {
	queuedAudioFileInfo := DequeueAudioInfo()
	lastAudioFileInfo := GetNowPlaying()
	log.Printf("Stopping %s\n", lastAudioFileInfo.Title)
	// mix.FadeOutMusic(fadeTime)

	nowPlaying = *queuedAudioFileInfo
	log.Printf("Starting %s\n", nowPlaying.Title)
	mus, err := mix.LoadMUS(nowPlaying.Filename)
	if err != nil {
		log.Printf("Error loading music file %s\n", nowPlaying.Filename)
	}
	mus.Play(-1)

	log.Printf("Now playing: %s\n", nowPlaying.Title)
	addToLog(lastAudioFileInfo)
	log.Println("Log: ", audioLog)
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
	err = mix.OpenAudio(mix.DEFAULT_FREQUENCY, mix.DEFAULT_FORMAT, mix.DEFAULT_CHANNELS, mix.DEFAULT_CHUNKSIZE)
	if err != nil {
		panic(err)
	}
}
