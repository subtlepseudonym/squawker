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

func Teardown() {
	vlcPlayer.Stop()
	mediaList.Release()
	vlcPlayer.Release()
}

func statusJSON(status int, message string) []byte {
	return []byte(fmt.Sprintf(`{"status":"%s","msg":"%s"}`, http.StatusText(status), message))
}

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
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(statusJSON(http.StatusInternalServerError, "File creation failed"))
		return
	}

	err = videoInfo.Download(audio, videoFile)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(statusJSON(http.StatusInternalServerError, "Audio retrieval failed"))
		videoFile.Close()
		err = os.Remove(filePath)
		if err != nil {
			log.Printf("ERR: %s\n", err.Error())
		}
		return
	}
	defer videoFile.Close()

	err = mediaList.AddMediaFromPath(filePath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(statusJSON(http.StatusInternalServerError, "MediaList add failed"))
		return
	}

	if !vlcPlayer.IsPlaying() {
		err = vlcPlayer.Play()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(statusJSON(http.StatusInternalServerError, "Play failed"))
			return
		}
	}

	log.Printf("ADD [%s | %s]\n", videoID, videoInfo.Title)
	w.Write(statusJSON(http.StatusOK, fmt.Sprintf("Added %s", videoInfo.Title)))
}

func NextHandler(w http.ResponseWriter, r *http.Request) {
	vlcPlayer.PlayNext()
	w.Write(statusJSON(http.StatusOK, "Playing next"))
}
