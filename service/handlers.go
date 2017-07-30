package service

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	serv "github.com/subtlepseudonym/go-utils/http"
)

const youtubeVideoInfoEndpoint string = "https://www.youtube.com/get_video_info/"
const videoIdRegex string = `[[:word:]-]{11}`

// TODO: add a logger and get to it

func NextHandler(w http.ResponseWriter, r *http.Request) {
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

	// TODO: toss this info into a queue, then do this just in time
	rawVideoInfo, err := getVideoInfo(videoId)
	if err != nil || rawVideoInfo == "" {
		serv.SimpleHttpResponse(w, http.StatusInternalServerError, "There was an error getting video info")
		return
	}

	// TODO: remember to do some logging with bitrate / codec etc
	audioInfo, err := parseAudioInfo(rawVideoInfo)
	if err != nil {
		serv.SimpleHttpResponse(w, http.StatusInternalServerError, "There was an error parsing video info")
		return
	}

	serv.SimpleHttpResponse(w, http.StatusOK, fmt.Sprintf("%+v", audioInfo))
}

func getVideoInfo(videoId string) (string, error) {
	// Make a call to youtube's get_video_info service
	res, err := http.Get(fmt.Sprintf(`%s?video_id=%s`, youtubeVideoInfoEndpoint, videoId))
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// Function name doesn't make a lot of sense in terms of flow
// This is because we're only interested in grabbing the audio stream
// For a bit more insight, read the comments below
func parseAudioInfo(rawVideoInfo string) (map[string][]string, error) {
	// Now that we've retrieved the video info, time to parse it
	videoInfoMap, err := url.ParseQuery(rawVideoInfo)
	if err != nil {
		return nil, err
	}

	// FIXME: need to do error checking
	// When videos are provided by VEVO or Sony Music Entertainment, they aren't retrievable this way

	// The 'adaptive_fmts' value is an array that always has length of 1
	audioStreamsInfo := strings.Split(videoInfoMap["adaptive_fmts"][0], ",")
	var vorbis map[string][]string
	var mp4a map[string][]string
	for _, streamInfo := range audioStreamsInfo {
		query, err := url.ParseQuery(streamInfo)
		if err != nil {
			// FIXME: this might count as laziness
			continue
		}
		// FIXME: there has to be a more efficient way of doing this
		// I'm preferring ogg vorbis because it tends to have to the highest bitrate
		// and I know that sdl2 will play it, with a fallback to mp4a if vorbis isn't available
		if strings.Contains(query["type"][0], "vorbis") {
			vorbis = query
		} else if strings.Contains(query["type"][0], "mp4a") {
			mp4a = query
		}
	}

	if vorbis != nil {
		return vorbis, nil
	} else if mp4a != nil {
		return mp4a, nil
	} else {
		// TODO: figure out how to play ogg opus files
		// worst case, I could just decode and re-encode them as vorbis ? (need to do some research)
		return nil, fmt.Errorf("No supported tracks available")
	}
}

func getAudioTrack(url string) {

}
