package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	configPath := os.Args[1]
	vintedConfigPath := filepath.Join(configPath, "vinted.json")
	generator := FeedGenerator{
		Providers: []ItemProvider{
			NewVintedProvider(vintedConfigPath),
		},
	}

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
