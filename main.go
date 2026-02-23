package main

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"os"
)

func main() {
	configPath := os.Args[1]
	generator := NewFeedGenerator(
		configPath,
		[]ItemProvider{
			NewVintedProvider(),
			NewBlocketProvider(),
		},
	)

	rssHandler := func(w http.ResponseWriter, req *http.Request) {
		dump, err := httputil.DumpRequest(req, false)
		if err != nil {
			slog.Error("Error dumping request", "error", err)
		} else {
			slog.Info("Got request", "request", dump)
		}

		data, err := generator.GetFeed()
		if err != nil {
			slog.Error("Error when generating feed", "error", err)
		}
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(200)
		w.Write(data)
	}

	http.HandleFunc("/", rssHandler)
	http.ListenAndServe(":8080", nil)
}
