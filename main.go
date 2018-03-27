package main

import (
	// "fmt"
	"log"
	"net/http"
	// "time"

	"squawker/music"
)

func main() {
	defer music.Teardown()

	http.HandleFunc("/add", music.AddHandler)
	http.HandleFunc("/next", music.NextHandler)

	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":15567", nil))
}
