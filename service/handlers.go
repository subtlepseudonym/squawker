package service

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	serv "github.com/subtlepseudonym/go-utils/http"

	"squawker/util"
)

const videoIdRegex string = `[[:word:]-]{11}`

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

	audioStreams, title, err := getAudioStreams(videoId)
	if err != nil || audioStreams == nil {
		serv.SimpleHttpResponse(w, http.StatusInternalServerError, "There was an error getting audio stream info")
		return
	}

	// TODO: remember to do some logging with bitrate / codec etc
	audioInfo, err := parseAudioInfo(audioStreams)
	if err != nil {
		serv.SimpleHttpResponse(w, http.StatusInternalServerError, "There was an error parsing audio stream info")
		return
	}

	bitrate, err := strconv.Atoi(audioInfo["bitrate"][0])
	if err != nil {
		serv.SimpleHttpResponse(w, http.StatusInternalServerError, "Received unparseable bitrate value from youtube video info")
		return
	}
	length, err := strconv.Atoi(audioInfo["clen"][0])
	if err != nil {
		serv.SimpleHttpResponse(w, http.StatusInternalServerError, "Received unparseable length value from youtube video info")
		return
	}

	toQueue := util.QueuedAudioInfo{
		AudioFileInfo: &util.AudioFileInfo{
			Id:       videoId,
			Filename: title,
			Bitrate:  bitrate,
			Length:   length,
			Stored:   false,
		},
		Url: audioInfo["url"][0],
	}

	serv.SimpleHttpResponse(w, http.StatusAccepted, fmt.Sprintf("%+v", toQueue.AudioFileInfo))
	util.EnqueueAudioInfo(&toQueue)
}
