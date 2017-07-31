package service

import (
	"bytes"
	"os/exec"
	"strings"
)

func getVideoTitle(videoId string) (string, error) {
	getTitle := exec.Command("youtube-dl", "-e", videoId)
	var out bytes.Buffer
	getTitle.Stdout = &out
	err := getTitle.Run()
	if err != nil {
		return "", err
	}

	return strings.Trim(out.String(), "\n"), nil
}
