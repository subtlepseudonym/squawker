package service

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strings"

	serv "github.com/subtlepseudonym/go-utils/http"

	"squawker/util"
)

const videoIdRegex string = `[[:word:]-]{11}`

var durationRegex *regexp.Regexp
var playNextChan chan bool

// TODO: add a logger and get to it

func AddHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		serv.SimpleHttpResponse(w, http.StatusMethodNotAllowed, fmt.Sprintf("Method %s not allowed.  Use GET", r.Method))
	}
	r.ParseForm()

	videoIdArr := r.Form["video"]
	if videoIdArr == nil {
		serv.SimpleHttpResponse(w, http.StatusBadRequest, "The 'video' parameter is required")
		return
	}

	videoId := videoIdArr[0] // we're only interested in the first video id
	matched, err := regexp.Match(videoIdRegex, []byte(videoId))
	if err != nil {
		serv.SimpleHttpResponse(w, http.StatusInternalServerError, "There was an error verifying video id")
		return
	}
	if !matched {
		serv.SimpleHttpResponse(w, http.StatusBadRequest, fmt.Sprintf(`The 'video' parameter must match "%s"`, videoIdRegex))
		return
	}
	log.Printf("ADD: %s\n", videoId)

	title, err := getVideoTitle(videoId)
	if err != nil {
		serv.SimpleHttpResponse(w, http.StatusInternalServerError, "There was an error getting video title")
		return
	}

	toQueue := util.AudioFileInfo{
		Id:     videoId,
		Title:  title,
		Stored: false,
	}

	serv.SimpleHttpResponse(w, http.StatusAccepted, fmt.Sprintf("%s", toQueue.Title))
	go util.EnqueueAudioInfo(&toQueue)
}

func NextHandler(w http.ResponseWriter, r *http.Request) {
	serv.SimpleHttpResponse(w, http.StatusOK, "Attempting to play next audio file")
	// TODO: find a way to assess what is / will block and send apropriate response to client
	go func() {
		util.PlayNext()
		// This is an overflowing channel, if chan is full, continue
		select {
		case playNextChan <- true:
		default:
		}
	}()
}

// TODO: time remaining handler (if for no other reason than testing)

func GetPlayNextChan() chan bool {
	return playNextChan
}

func getVideoTitle(videoId string) (string, error) {
	getTitle := exec.Command("youtube-dl", "-e", "--", videoId)
	var out bytes.Buffer
	getTitle.Stdout = &out
	err := getTitle.Run()
	if err != nil {
		return "", err
	}

	return strings.Trim(out.String(), "\n"), nil
}

func init() {
	playNextChan = make(chan bool, 1)
}
