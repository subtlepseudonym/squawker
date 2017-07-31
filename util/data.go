package util

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/veandco/go-sdl2/mix"
)

const PreferredFormat string = "mp3"
const preferredQuality string = "0" // from 0-9 where 0 is best
const defaultTemplate string = `/%(id)s.%(ext)s`

type AudioFileInfo struct {
	Id       string // this is the id from youtube
	Filename string
	Title    string
	Mus      *mix.Music
	Stored   bool // is file currently downloaded
}

var dlQueue chan *AudioFileInfo // queue of things to download, buffer size of 1
var queue chan *AudioFileInfo   // queue of things to play, is buffered
var nowPlaying AudioFileInfo    // this var is manipulated from music.go FYI
var audioLog []AudioFileInfo    // treated like a queue, oldest files are deleted

func EnqueueAudioInfo(audioFileInfo *AudioFileInfo) {
	dlQueue <- audioFileInfo
	log.Printf("DL: starting %s\n", audioFileInfo.Id)
	err := downloadAudioFile(audioFileInfo)
	if err != nil {
		<-dlQueue // FIXME: this is just lazy
		return
	}
	queue <- <-dlQueue // this is the dumbest looking syntax
}

func downloadAudioFile(queuedAudioFileInfo *AudioFileInfo) error {
	dlCmd := exec.Command("youtube-dl", "-x", "--audio-format", PreferredFormat, "--audio-quality", preferredQuality, "-o", GetAudioFileDirectory()+defaultTemplate, "--", queuedAudioFileInfo.Id)
	var out bytes.Buffer
	dlCmd.Stdout = &out
	err := dlCmd.Run()
	if err != nil {
		return err
	}

	log.Printf("DL: complete %s\n", queuedAudioFileInfo.Id)
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

func GetPrettyAudioLogString(limit int) string {
	var buf bytes.Buffer
	buf.WriteString("-   ")
	for i := 0; i < limit; i++ {
		fileInfo, err := GetFromLog(i)
		if err != nil {
			break
		}
		buf.WriteString(fileInfo.Id)
		buf.WriteString(" - ")
		buf.WriteString(fileInfo.Title)
		buf.WriteString("\n-   ")
	}
	return buf.String()
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
