package util

import (
	"fmt"
	"os"
)

type AudioFileInfo struct {
	Id       string // this is the id from youtube
	Filename string
	Bitrate  int
	Length   int  // length in seconds
	Stored   bool // is file currently downloaded
}

type QueuedAudioInfo struct {
	*AudioFileInfo
	Url string // TODO: this might cause problems if the links expire (hopefully an edge case)
}

var queued chan *QueuedAudioInfo // only downside is that we can't view the queue unless I store that info elsewhere
var nowPlaying AudioFileInfo
var audioLog []AudioFileInfo // treated like a queue, oldest files are deleted

func EnqueueAudioInfo(audioInfo *QueuedAudioInfo) {
	queued <- audioInfo
}

func DequeueAudioInfo() *QueuedAudioInfo {
	return <-queued
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
	err := os.MkdirAll(GetAudioFileDirectory(), os.ModeDir)
	if err != nil {
		panic(err) // FIXME
	}

	queued = make(chan *QueuedAudioInfo, GetQueueSize())
}
