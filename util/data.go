package util

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type AudioFileInfo struct {
	Id       string // this is the id from youtube
	Filename string
	Bitrate  int
	Length   int  // length in seconds
	Stored   bool // is file currently downloaded
	Url      string
}

var dlQueue chan *AudioFileInfo // queue of things to download, buffer size of 1
var queue chan *AudioFileInfo   // queue of things to play, is buffered
var nowPlaying AudioFileInfo    // this var is manipulated from music.go FYI
var audioLog []AudioFileInfo    // treated like a queue, oldest files are deleted

func EnqueueAudioInfo(audioFileInfo *AudioFileInfo) {
	dlQueue <- audioFileInfo
	log.Printf("Downloading %s\n", audioFileInfo.Id)
	downloadAudioFile(audioFileInfo)
	queue <- <-dlQueue // this is the dumbest looking syntax
}

func downloadAudioFile(queuedAudioFileInfo *AudioFileInfo) {
	if queuedAudioFileInfo == nil {
		// TODO: just keep the queue full with suggested tracks or something similar
		return // need to wait until another song is added
	}

	res, err := http.Get(queuedAudioFileInfo.Url)
	if err != nil {
		// TODO: log, this is really not good, skip song?
		return
	}

	log.Println("Creating file")
	newAudioFile, err := os.Create(fmt.Sprintf("%s/%s", GetAudioFileDirectory(), queuedAudioFileInfo.Filename))
	if err != nil {
		// TODO: log, this is not good, likely filename is already taken, prune old files?
		return
	}
	log.Println("File created")

	defer res.Body.Close()
	_, err = io.Copy(newAudioFile, res.Body)
	if err != nil {
		// TODO: log, skip to next
		return
	}
	queuedAudioFileInfo.Stored = true
	log.Printf("Downloaded %s\n", queuedAudioFileInfo.Id)
}

func DequeueAudioInfo() *AudioFileInfo {
	return <-queue
}

func GetNowPlaying() AudioFileInfo {
	return nowPlaying
}

func GetFromLog(idx int) (*AudioFileInfo, error) {
	if len(audioLog) <= idx {
		return nil, fmt.Errorf("Index out of bounds: requested %d, length of log is %d", idx, len(audioLog))
	}
	return &audioLog[idx], nil
}

func addToLog(a AudioFileInfo) {
	audioLog = append([]AudioFileInfo{a}, audioLog...)
	if len(audioLog) > GetLogSize() {
		audioLog = audioLog[:GetLogSize()]
	}

	// Avoiding index out of bounds
	if len(audioLog) < GetFileBacklogSize() {
		return
	}
	fileInfoToPrune := audioLog[GetFileBacklogSize()]
	if fileInfoToPrune.Stored {
		err := os.Remove(fileInfoToPrune.Filename)
		if err != nil {
			// TODO: log the issue (this is fairly serious)
			// mark that a backlogCheck is needed
			return
		}
		fileInfoToPrune.Stored = false
	}
}

// Depending upon the size of GetFileBacklogSize(), this can get expensive (basically it's garbage collection)
// In most cases this will never need to be run, but it certainly makes me feel better knowing the functionality exists :D
func backlogCheck() {
	for i, fileInfo := range audioLog {
		if i < GetFileBacklogSize() {
			continue
		}

		if fileInfo.Stored {
			err := os.Remove(fileInfo.Filename)
			if err != nil {
				// TODO: this is very serious, log and maybe start screaming
				continue
			}
			fileInfo.Stored = false
		} else {
			break
		}
	}
}

func init() {
	err := os.MkdirAll(GetAudioFileDirectory(), 0777)
	if err != nil {
		panic(err) // FIXME
	}

	queue = make(chan *AudioFileInfo, GetQueueSize())
	dlQueue = make(chan *AudioFileInfo, 1)
}
