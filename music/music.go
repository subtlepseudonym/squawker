package music

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/adrg/libvlc-go"
	"github.com/pkg/errors"
	"github.com/rylio/ytdl"
)

const videoIDRegex string = `[[:word:]-]{11}`
const audioFileDir string = "audio_files"

var vlcPlayer *vlc.ListPlayer
var audioCache = make(map[string]Audio)

type Audio struct {
	VideoID  string
	VideoURL string
	Title    string
}

func init() {
	// --mod-megabass
	// --adaptive-logic {,predictive,nearoptimal,rate,fixedrate,lowest,highest}
	err := vlc.Init("--no-video", "-q") //, "--sout-mp4-faststart")
	if err != nil {
		panic(err)
	}

	vlcPlayer, err = vlc.NewListPlayer()
	if err != nil {
		panic(err)
	}

	mediaList, err := vlc.NewMediaList()
	if err != nil {
		panic(err)
	}

	err = vlcPlayer.SetMediaList(mediaList)
	if err != nil {
		panic(err)
	}
}

// Teardown free memory allocated by cgo in the vlc package
// This function should be called before exiting
func Teardown() {
	err := vlcPlayer.Stop()
	if err != nil {
		log.Printf("vlcPlayer.Stop failed: %s\n", err)
	}
	err = vlcPlayer.MediaList().Release()
	if err != nil {
		log.Printf("mediaList.Release failed: %s\n", err)
	}
	err = vlcPlayer.Release()
	if err != nil {
		log.Printf("vlcPlayer.Release failed: %s\n", err)
	}
}

// statusJSON is just a nice abstraction for marshalling simply JSON objects
func statusJSON(status int, message string) []byte {
	return []byte(fmt.Sprintf(`{"status":"%s","msg":%s}`, http.StatusText(status), strconv.Quote(message)))
}

// getAudio checks the cache, then retrieves video metadata and returns and Audio struct
func getAudio(videoID string) (*Audio, error) {
	if a, ok := audioCache[videoID]; ok {
		return &a, nil
	}

	videoInfo, err := ytdl.GetVideoInfoFromID(videoID)
	if err != nil {
		return nil, errors.Wrap(err, "get video from ID failed")
	}

	var audio ytdl.Format
	for _, format := range videoInfo.Formats.Best(ytdl.FormatAudioEncodingKey) {
		formatStr, ok := format.ValueForKey("type").(string)
		if !ok {
			log.Printf("value assertion error: %v", format.ValueForKey("type"))
			continue
		}
		if strings.Contains(formatStr, "audio/") {
			audio = format
			break
		}
	}

	audioURL, err := videoInfo.GetDownloadURL(audio)
	if err != nil {
		return nil, errors.Wrap(err, "get audio URL failed")
	}

	a := Audio{
		VideoID:  videoID,
		VideoURL: audioURL.String(),
		Title:    videoInfo.Title,
	}

	audioCache[a.VideoID] = a
	return &a, nil
}

// AddHandler uses the ytdl and vlc packages to download requested video's audio tracks
// and add them to the playlist
func AddHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write(statusJSON(http.StatusMethodNotAllowed, "Only 'GET' allowed"))
		return
	}
	r.ParseForm()

	videoIDArr := r.Form["v"]
	if videoIDArr == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(statusJSON(http.StatusBadRequest, "The 'v' parameter is required"))
		return
	}

	videoID := videoIDArr[0] // FIXME: should accept the whole array
	matched, err := regexp.MatchString(videoIDRegex, videoID)
	if err != nil {
		log.Printf("regex matching failed: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(statusJSON(http.StatusInternalServerError, fmt.Sprintf("There was an error verifying video ID %s", videoID)))
		return
	}
	if !matched {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(statusJSON(http.StatusBadRequest, fmt.Sprintf("VideoID '%s' does not match regex '%s'", videoID, videoIDRegex)))
		return
	}

	audio, err := getAudio(videoID)
	if err != nil {
		log.Printf("get audio failed: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(statusJSON(http.StatusInternalServerError, "Get audio failed"))
		return
	}

	vlcPlayer.MediaList().Lock()
	err = vlcPlayer.MediaList().AddMediaFromURL(audio.VideoURL)
	if err != nil {
		log.Printf("add to media list failed: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(statusJSON(http.StatusInternalServerError, "MediaList add failed"))
		vlcPlayer.MediaList().Unlock()
		return
	}
	vlcPlayer.MediaList().Unlock()

	if !vlcPlayer.IsPlaying() {
		time.Sleep(1 * time.Second)
		err = vlcPlayer.Play()
		if err != nil {
			log.Printf("play failed: %s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(statusJSON(http.StatusInternalServerError, "Play failed"))
			return
		}
	}

	log.Printf("ADD [%s | %s]\n", audio.VideoID, audio.Title)
	w.Write(statusJSON(http.StatusOK, fmt.Sprintf("Added %s", audio.Title)))
}

// NextHandler attempts to play the next song
func NextHandler(w http.ResponseWriter, r *http.Request) {
	err := vlcPlayer.PlayNext()
	if err != nil {
		log.Printf("play next failed: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(statusJSON(http.StatusInternalServerError, "Play next failed"))
		return
	}
	w.Write(statusJSON(http.StatusOK, "Playing next"))
}

// PrevHandler attempts to play the previous song
func PrevHandler(w http.ResponseWriter, r *http.Request) {
	err := vlcPlayer.PlayPrevious()
	if err != nil {
		log.Printf("play prev failed: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(statusJSON(http.StatusInternalServerError, "Play prev failed"))
		return
	}
	w.Write(statusJSON(http.StatusOK, "Playing prev"))
}

// ToggleHandler attempts to toggle the ListPlayer's pause state
func ToggleHandler(w http.ResponseWriter, r *http.Request) {
	err := vlcPlayer.TogglePause()
	if err != nil {
		log.Printf("toggle pause failed: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(statusJSON(http.StatusInternalServerError, "Toggle pause failed"))
		return
	}
	time.Sleep(250 * time.Millisecond)
	pausedStr := "paused"
	if vlcPlayer.IsPlaying() {
		pausedStr = "playing"
	}
	w.Write(statusJSON(http.StatusOK, fmt.Sprintf("Now %s", pausedStr)))
}
