package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	configPath := os.Args[1]
	generator := NewFeedGenerator(
		configPath,
		[]ItemProvider{
			NewVintedProvider(),
		},
	)

	rssHandler := func(w http.ResponseWriter, req *http.Request) {
		data, err := generator.GetFeed()
		if err != nil {
			log.Println("Error", err)
		}
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(200)
		w.Write(data)
	}

	http.HandleFunc("/", rssHandler)
	http.ListenAndServe(":8080", nil)
}
