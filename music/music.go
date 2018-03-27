package music

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/adrg/libvlc-go"
	"github.com/rylio/ytdl"
	"github.com/satori/go.uuid"
)

const videoIDRegex string = `[[:word:]-]{11}`
const audioFileDir string = "audio_files"

var vlcPlayer *vlc.ListPlayer
var mediaList *vlc.MediaList

func init() {
	err := vlc.Init("--no-video")
	if err != nil {
		panic(err) // FIXME
	}

	vlcPlayer, err = vlc.NewListPlayer()
	if err != nil {
		panic(err)
	}

	mediaList, err = vlc.NewMediaList()
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
	err = mediaList.Release()
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
	return []byte(fmt.Sprintf(`{"status":"%s","msg":"%s"}`, http.StatusText(status), message))
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

	videoIDArr := r.Form["video"]
	if videoIDArr == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(statusJSON(http.StatusBadRequest, "The 'video' parameter is required"))
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

	videoInfo, err := ytdl.GetVideoInfoFromID(videoID)
	if err != nil {
		log.Printf("get video from ID failed: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(statusJSON(http.StatusInternalServerError, "Info retrieval failed"))
		return
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

	filePath := fmt.Sprintf("%s/%s.%s", audioFileDir, uuid.NewV4().String(), audio.Extension)
	videoFile, err := os.Create(filePath)
	if err != nil {
		log.Printf("file creation failed: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(statusJSON(http.StatusInternalServerError, "File creation failed"))
		return
	}

	err = videoInfo.Download(audio, videoFile)
	if err != nil {
		log.Printf("audio download failed: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(statusJSON(http.StatusInternalServerError, "Audio retrieval failed"))
		videoFile.Close()
		err = os.Remove(filePath)
		if err != nil {
			log.Printf("remove file failed: %s\n", err)
		}
		return
	}
	defer videoFile.Close()

	err = mediaList.AddMediaFromPath(filePath)
	if err != nil {
		log.Printf("add to media list failed: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(statusJSON(http.StatusInternalServerError, "MediaList add failed"))
		return
	}

	if !vlcPlayer.IsPlaying() {
		err = vlcPlayer.Play()
		if err != nil {
			log.Printf("play failed: %s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(statusJSON(http.StatusInternalServerError, "Play failed"))
			return
		}
	}

	log.Printf("ADD [%s | %s]\n", videoID, videoInfo.Title)
	w.Write(statusJSON(http.StatusOK, fmt.Sprintf("Added %s", videoInfo.Title)))
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
	pausedStr := "paused"
	if vlcPlayer.IsPlaying() {
		pausedStr = "playing"
	}
	w.Write(statusJSON(http.StatusOK, fmt.Sprintf("Now %s", pausedStr)))
}
