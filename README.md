### Squawker ###

This intention of this project is to the solve the 'too many DJs' problem.  Anytime you've got more than one musically inclined friend over and music is playing, they all want to pick the next song.

Squawker provides a REST endpoint for adding songs, sourced directly from youtube, to a concurrency safe queue which are then played via vlc.  I'm working on reducing the number of external (non-golang) with the ultimate goal of providing a standalong fat binary deployment.

You can build squawker from source with 
```bash
GOOS=linux GOARCH=amd64 go build -a -o "squawker-$(git describe --abbrev=0 --tags).linux.amd64" main.go
```
This command assumes that you're building for 64-bit linux and that you've fetched this repo's version tags.
