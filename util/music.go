package util

import (
	"log"

	"github.com/adrg/libvlc-go"
)

var vlcPlayer *vlc.Player
var mediaList *vlc.MediaList

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
		lastAudioFileInfo.Media.Release()
		addToLog(*lastAudioFileInfo)
	}
	nowPlaying = queuedAudioFileInfo

	media, err := vlc.NewMediaFromPath(nowPlaying.Filename)
	if err != nil {
		log.Printf("Error loading media file ( %s ): %s\n", nowPlaying.Filename, err.Error())
		return
	}
	nowPlaying.Media = media

	err = vlcPlayer.SetMedia(media)
	if err != nil {
		// TODO: try to play next song in the queue
		log.Printf("Error setting media: %s\n", err.Error())
		return
	}
	err = vlcPlayer.Play()
	if err != nil {
		log.Printf("Error playing loaded media file: %s\n", err.Error())
		return
	}

	log.Printf("Now playing: %s\n", nowPlaying.Title)
	log.Println("Log --", GetPrettyAudioLogString(5))

	playNextAlreadyQueued = false
}

func GetPlayer() *vlc.Player {
	return vlcPlayer
}

func CleanUp() {
	// Will be deferred in main()
	vlcPlayer.Stop()
	mediaList.Release()
	vlcPlayer.Release()
}

func PlayerIsPlaying() bool {
	return vlcPlayer.IsPlaying()
}

func GetMedia() *vlc.Media {
	return vlcPlayer.Media()
}

func GetMediaLength() int {
	length, err := vlcPlayer.MediaLength()
	if err != nil {
		log.Printf("Error getting media length: %s\n", err.Error())
		return 0 // FIXME: should probably fix this, but currently just using main.go logic to play next song
	}
	return length
}

func init() {
	err := vlc.Init("--no-video")
	if err != nil {
		panic(err) // FIXME
	}

	vlcPlayer, err = vlc.NewPlayer()
	if err != nil {
		panic(err)
	}
}
