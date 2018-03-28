### Squawker ###

This intention of this project is to the solve the 'too many DJs' problem.  Anytime you've got more than one musically inclined friend over and music is playing, they all want to pick the next song.

Squawker receives video IDs from a REST endpoint and adds them to a queue.  Next and previous commands are provided via canonical endpoints.  By default, Squawker begins playing anytime the queue is populated with an unplayed track and doesn't repeat the queue.

You can build squawker from source like so:
```bash
go build main.go
```

Personally, I use the following for more descriptive binary names:

```bash
GOOS=linux GOARCH=amd64 go build -a -o "squawker-$(git describe --abbrev=0 --tags).linux.amd64" main.go
```
This command assumes that you're building for 64-bit linux and that you've fetched this repo's version tags.


#### Example REST Calls ####

Songs can be added to the queue like so:
```bash
curl "localhost:15567/add?v=dQw4w9WgXcQ"
```

Playing the next song:
`curl "localhost:15567/next"`
Playing the previous song:
`curl "localhost:15567/prev"`
And toggling pause state:
`curl "localhost:15567/toggle"`


#### Notes ####

All told, I'm using vlc to playback the downloaded audio files.  This is something akin to doing pest control with a flamethrower (total overkill) and I may actually be able to just download the audio using vlc alone (I'm currently using another library for video info / dls).

I'm currently building with vlc v2.2.4 because the latest build of vlc for arch linux is v3.0.1 and it's got a bug where list_player instances don't play the next media.  I'll upgrade to v3.0.2 when it's available; the bug is fixed in that patch (https://www.videolan.org/developers/vlc-branch/NEWS)