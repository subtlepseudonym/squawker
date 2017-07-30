package service

import (
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/robertkrimen/otto"
	serv "github.com/subtlepseudonym/go-utils/http"
	xhtml "golang.org/x/net/html"
)

const youtubeVideoInfoEndpoint string = "https://www.youtube.com/get_video_info"
const youtubeDirectEndpoint string = "https://www.youtube.com/watch"
const videoIdRegex string = `[[:word:]-]{11}`
const targetScriptPrefix string = "var ytplayer"

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
	rawVideoInfo, err := getAudioStreams(videoId)
	if err != nil || rawVideoInfo == nil {
		serv.SimpleHttpResponse(w, http.StatusInternalServerError, "There was an error getting audio stream info")
		return
	}

	// TODO: remember to do some logging with bitrate / codec etc
	audioInfo, err := parseAudioInfo(rawVideoInfo)
	if err != nil {
		serv.SimpleHttpResponse(w, http.StatusInternalServerError, "There was an error parsing audio stream info")
		return
	}

	serv.SimpleHttpResponse(w, http.StatusOK, fmt.Sprintf("%+v", audioInfo))
}

func getAudioStreams(videoId string) ([]string, error) {
	audioStreams, err := getAudioStreamsFromYoutubeVideoInfoEndpoint(fmt.Sprintf(`%s?video_id=%s`, youtubeVideoInfoEndpoint, videoId))
	if err != nil {
		return nil, err
	}

	// TODO: log whether video info was retrieved optimistically
	// Returning nil audioStreams indicates that there is DRM
	if audioStreams != nil {
		return audioStreams, nil
	}

	audioStreams, err = getAudioStreamsFromYoutubeDirectly(fmt.Sprintf(`%s?v=%s`, youtubeDirectEndpoint, videoId))
	if err != nil {
		return nil, err
	}

	return audioStreams, nil
}

// Talk about title gore
// This makes the optimistic assumption that the video will not be VEVO / SME
func getAudioStreamsFromYoutubeVideoInfoEndpoint(endpoint string) ([]string, error) {
	res, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	videoInfoMap, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, err
	}

	if videoInfoMap["status"][0] != "fail" {
		// All good, no DRM here
		audioFmtsList := strings.Split(videoInfoMap["adaptive_fmts"][0], ",")
		return audioFmtsList, nil
	}
	// Indicate to getAudioStreams() that we need to get the stream info from youtube directly
	return nil, nil
}

func getAudioStreamsFromYoutubeDirectly(endpoint string) ([]string, error) {
	res, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}

	// vm will be our js parser
	// The code we're looking for references the window object, so it needs to be defined
	vm := otto.New()
	_, err = vm.Run(`window = {}`) // I would complain about how hacky this is if the whole project wasn't a hack
	if err != nil {
		return nil, err
	}

	// It
	defer res.Body.Close()
	tokenizer := xhtml.NewTokenizer(res.Body)
	for tokenizer.Next() != xhtml.ErrorToken {
		token := tokenizer.Token()
		if token.Data != "script" {
			// the code we're looking for is in a script tag (on line 226 last I checked)
			continue
		}

		tokenizer.Next() // token was the tag, now we need the text inside the script tag
		token = tokenizer.Token()
		if !strings.HasPrefix(token.String(), targetScriptPrefix) {
			// If this script content doesn't match our target script, keep moving
			continue
		}

		// The script contains a whole lot of html escaped sequences
		htmlParsedScript := html.UnescapeString(token.String())
		_, err = vm.Run(htmlParsedScript) // Running the target script so that ytplayer.config.args.adaptive_fmts will be retrievable
		if err != nil {
			return nil, err
		}
		// Some of the ugliest code I've ever written, but very robust ;)
		ytplayer, err := vm.Get("ytplayer")
		if err != nil {
			return nil, err
		}
		ytplayerConfig, err := ytplayer.Object().Get("config")
		if err != nil {
			return nil, err
		}
		ytplayerConfigArgs, err := ytplayerConfig.Object().Get("args")
		if err != nil {
			return nil, err
		}
		ytplayerConfigArgsAdaptiveFmts, err := ytplayerConfigArgs.Object().Get("adaptive_fmts")
		if err != nil {
			return nil, err
		}
		return strings.Split(ytplayerConfigArgsAdaptiveFmts.String(), ","), nil
	}

	return nil, fmt.Errorf("Unable to find audio stream info; Last token: %+v", tokenizer.Token())
}

func parseAudioInfo(audioStreamsInfo []string) (map[string][]string, error) {
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
