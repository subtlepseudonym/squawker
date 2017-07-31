package service

import (
	"bytes"
	"os/exec"
	"strings"
)

// This is defined in this package because it is only called from handlers.go
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
