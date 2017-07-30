package util

import (
	"log"
)

// This function assumes that the next QueuedAudioInfo has already been downloaded
func PlayNext() {
	queuedAudioFileInfo := DequeueAudioInfo()
	// TODO: stop nowPlaying if not already
	lastAudioFileInfo := GetNowPlaying()
	log.Printf("Stopping %s\n", lastAudioFileInfo.Id)
	nowPlaying = *queuedAudioFileInfo
	// TODO: start nowPlaying
	log.Printf("Starting %s\n", nowPlaying.Id)
	addToLog(lastAudioFileInfo)
	log.Printf("Now playing: %s\n", nowPlaying.Filename)
	log.Println("Log: ", audioLog)
}
