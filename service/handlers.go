package service

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	// "strconv"

	serv "github.com/subtlepseudonym/go-utils/http"

	"squawker/util"
)

const videoIdRegex string = `[[:word:]-]{11}`
const durationRegexStr string = `:??[0-9]{0,2}:??`

var durationRegex *regexp.Regexp

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
	log.Printf("AD: received %s\n", videoId)

	title, err := getVideoTitle(videoId)
	if err != nil {
		serv.SimpleHttpResponse(w, http.StatusInternalServerError, "There was an error getting video title")
		return
	}

	toQueue := util.AudioFileInfo{
		Id:       videoId,
		Filename: fmt.Sprintf("%s/%s%s", util.GetAudioFileDirectory(), videoId, util.GetFileExtension()),
		Title:    title,
		Stored:   false,
	}

	serv.SimpleHttpResponse(w, http.StatusAccepted, fmt.Sprintf("%s", toQueue.Title))
	go util.EnqueueAudioInfo(&toQueue)
}

func NextHandler(w http.ResponseWriter, r *http.Request) {
	serv.SimpleHttpResponse(w, http.StatusOK, "Attempting to play next audio file")
	go util.PlayNext()
}

// func parseDuration(dur string) int {
// 	matches := durationRegex.FindAllStringSubmatch(duration, -1)
// 	var matches_int []int
// 	for _, match := range matches {
// 		num, err :=
// 	}
// }

func init() {
	durationRegex = regexp.MustCompile(durationRegexStr)
}
