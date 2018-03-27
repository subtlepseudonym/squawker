package main

import (
	"log"
	"net/http"

	"squawker/music"
)

func main() {
	defer music.Teardown()

	http.HandleFunc("/add", music.AddHandler)
	http.HandleFunc("/next", music.NextHandler)
	http.HandleFunc("/prev", music.PrevHandler)
	http.HandleFunc("/toggle", music.ToggleHandler)

	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":15567", nil))
}
