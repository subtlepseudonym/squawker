package util

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/adrg/libvlc-go"
	"github.com/rylio/ytdl"
	"github.com/satori/go.uuid"
)

const preferredQuality string = "0" // from 0-9 where 0 is best

type AudioFileInfo struct {
	Id       string // this is the id from youtube
	Filename string
	Title    string
	Media    *vlc.Media
	Duration time.Duration
	Stored   bool // is file currently downloaded
}

func (a *AudioFileInfo) Equals(other *AudioFileInfo) bool {
	return a.Id == other.Id
}

var dlLock sync.Mutex         // ensures that only one file more than queueSize is downloaded
var queue chan *AudioFileInfo // queue of things to play, is buffered
var nowPlaying *AudioFileInfo // this var is manipulated from music.go FYI
var audioLog []AudioFileInfo  // treated like a ring buffer, oldest files are deleted

// FIXME: this has to be a better way to check if file is downloaded, id->downloaded map?
func EnqueueAudioInfo(audioFileInfo *AudioFileInfo) {
	dlLock.Lock()
	downloadedAudioFileInfo, err := downloadAudioFile(audioFileInfo)
	if err != nil {
		log.Println("ERR: ", err.Error())
		dlLock.Unlock()
		return
	}
	queue <- downloadedAudioFileInfo
	dlLock.Unlock()
}

// FIXME: The interaction with youtube-dl is not the most transparent thing in the world
func downloadAudioFile(queuedAudioFileInfo *AudioFileInfo) (*AudioFileInfo, error) {
	fileId := uuid.NewV4()
	vid, err := ytdl.GetVideoInfoFromID(queuedAudioFileInfo.Id)
	if err != nil {
		return nil, err
	}

	bestAudio := vid.Formats.Best(ytdl.FormatAudioEncodingKey)
	var audio ytdl.Format
	for _, format := range bestAudio {
		if strings.Contains(format.ValueForKey("type").(string), "audio/") {
			audio = format
			break
		}
	}

	filePath := fmt.Sprintf("%s/%s.%s", GetAudioFileDirectory(), fileId.String(), audio.Extension)
	vidFile, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer vidFile.Close()

	err = vid.Download(audio, vidFile)
	if err != nil {
		return nil, err
	}

	queuedAudioFileInfo.Filename = filePath
	queuedAudioFileInfo.Stored = true
	return queuedAudioFileInfo, nil
}

func DequeueAudioInfo() *AudioFileInfo {
	return <-queue
}

func GetNowPlaying() *AudioFileInfo {
	return nowPlaying
}

func GetFromLog(idx int) (*AudioFileInfo, error) {
	if len(audioLog) <= idx {
		return nil, fmt.Errorf("Index out of bounds: requested %d, length of log is %d", idx, len(audioLog))
	}
	return &audioLog[idx], nil
}

// Doesn't contain a newline at the end of the string (use with Println)
func GetPrettyAudioLogString(limit int) string {
	var buf bytes.Buffer
	for i, log := range audioLog {
		if i >= limit {
			break
		}
		buf.WriteString("\n")
		buf.WriteString("-   ")
		buf.WriteString(log.Id)
		buf.WriteString(" | ")
		buf.WriteString(log.Title)
	}
	buf.WriteString("\n")
	return buf.String()
}

// FIXME: if song is added twice and the additions are more than queueSize apart, the second add will get file not found
func addToLog(a AudioFileInfo) {
	audioLog = append([]AudioFileInfo{a}, audioLog...)
	if len(audioLog) > GetLogSize() {
		audioLog = audioLog[:GetLogSize()]
	}

	// Avoiding index out of bounds
	if len(audioLog) < GetNumFilesMaintained() {
		return
	}
	fileInfoToPrune := audioLog[GetNumFilesMaintained()-1]
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

// TODO: add functionality to remove songs found in directory that aren't in audioLog, or are and aren't marked stored
// Depending upon the size of GetNumFilesMaintained(), this can get expensive (basically it's garbage collection)
// In most cases this will never need to be run, but it certainly makes me feel better knowing the functionality exists :D
func backlogCheck() {
	for i, fileInfo := range audioLog {
		if i < GetNumFilesMaintained() {
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
}
