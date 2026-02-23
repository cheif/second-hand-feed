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

		baseURL := *req.URL
		proto := req.Header.Get("x-Forwarded-Proto")
		if proto == "" {
			proto = "http"
		}
		baseURL.Host = req.Host
		data, err := generator.GetFeed(baseURL)
		if err != nil {
			slog.Error("Error when generating feed", "error", err)
		}
		w.Header().Set("Content-Type", "application/atom+xml")
		w.WriteHeader(200)
		w.Write(data)
	}

	http.HandleFunc("/", rssHandler)
	http.ListenAndServe(":8080", nil)
}
