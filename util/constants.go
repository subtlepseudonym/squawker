package util

import (
	"flag"
	"os"
)

const (
	// quiet (almost no logging), default/normal (log significant stuff), verbose (log everything)
	VerbosityQuiet = iota
	VerbosityNormal
	VerbosityVerbose
)

const defaultVerbosity = VerbosityNormal
const defaultPort = 15567
const defaultQueueSize = 2                 // size of queue waiting to be played, number of downloaded files is at most (queueSize + numFilesMaintained + 1)
const defaultLogSize = 25                  // how many AudioFileInfo objects to keep after they've been played
const defaultNumFilesMaintained = 5        // maintain the last 5 songs played, delete anything older
const defaultMaxSongLength = 600           // length in seconds
const defaultFileDirectory = "audio_files" // just the name of the folder we're keeping audio files in

var verbosity int
var port int
var queueSize int
var logSize int
var numFilesMaintained int
var maxSongLength int
var fileDirectory string

func GetVerbosity() int {
	return verbosity
}

func GetPort() int {
	return port
}

func GetQueueSize() int {
	return queueSize
}

func GetLogSize() int {
	return logSize
}

func GetFileBacklogSize() int {
	return numFilesMaintained
}

func GetMaxSongLength() int {
	return maxSongLength
}

func GetAudioFileDirectory() string {
	return fileDirectory
}

func init() {
	var q, v, vv bool
	flag.BoolVar(&q, "q", false, "Almost no logging")
	flag.BoolVar(&v, "v", true, "Log significant interactions - this overrides q")
	flag.BoolVar(&vv, "vv", false, "Log everything - this overrides q and v")

	flag.IntVar(&port, "port", defaultPort, "Set the port where the server listens for incoming connections")
	flag.IntVar(&queueSize, "queue-size", defaultQueueSize, "Number of audio files to store in queue before playing (larger number means more memory, but fewer blocking goroutines)")
	flag.IntVar(&logSize, "log-size", defaultLogSize, "Number of audio files to store in log (so they can be replayed)")
	flag.IntVar(&numFilesMaintained, "num-files", defaultNumFilesMaintained, "Only this many files will be stored with oldest files being deleted first")
	flag.IntVar(&maxSongLength, "max-length", defaultMaxSongLength, "This sets the maximum song length allowed (in seconds)")
	flag.StringVar(&fileDirectory, "audio-dir", defaultFileDirectory, "This sets the name of the directory where audio files are stored")

	flag.Parse()

	// Tougher to read, but easier to use
	if vv {
		verbosity = VerbosityVerbose
	} else if v {
		verbosity = VerbosityNormal
	} else if q {
		verbosity = VerbosityQuiet
	}

	// Making fileDirectory absolute
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fileDirectory = wd + "/" + fileDirectory
}
