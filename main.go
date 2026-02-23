package main

import (
	"html/template"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
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

		if req.URL.Path != "/" {
			w.WriteHeader(404)
			return
		}

		accept := req.Header.Get("Accept")
		shouldServeFeed := strings.Contains(accept, "application/atom+xml") || (strings.Contains(accept, "*/*") && !strings.Contains(accept, "text/html"))

		if shouldServeFeed {
			// Should be able to handle our feed
			data, err := generator.GetFeed(baseURL)
			if err != nil {
				slog.Error("Error when generating feed", "error", err)
			}
			w.Header().Set("Content-Type", "application/atom+xml")
			w.WriteHeader(200)
			w.Write(data)
		} else {
			// Serve HTML
			tmpl, err := template.ParseFiles("templates/index.html")
			if err != nil {
				slog.Error("Error when parsing template", "error", err)
			}
			queries, err := generator.GetQueries()
			if err != nil {
				slog.Error("Error when fetching queries", "error", err)
			}
			tmpl.Execute(w, queries)
		}
	}

	http.HandleFunc("POST /queries/add", func(w http.ResponseWriter, req *http.Request) {
		dump, err := httputil.DumpRequest(req, false)
		if err != nil {
			slog.Error("Error dumping request", "error", err)
		} else {
			slog.Info("Got request", "request", dump)
		}
		_ = req.ParseForm()
		query := req.Form.Get("query")
		slog.Info(query)
		if query == "" {
			w.WriteHeader(400)
		} else {
			err := generator.AddQuery(query)
			if err != nil {
				slog.Error("Error creating query", "error", err)
				w.WriteHeader(400)
			} else {
				w.Header().Set("Location", "/")
				w.WriteHeader(201)
			}
		}
	})

	http.HandleFunc("/", rssHandler)
	http.ListenAndServe(":8080", nil)
}
