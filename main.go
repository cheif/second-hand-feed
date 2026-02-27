package main

import (
	"html/template"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/cheif/second-hand-feed/providers"
)

func main() {
	configPath := os.Args[1]
	generator := NewFeedGenerator(
		configPath,
		[]providers.ItemProvider{
			providers.NewVintedProvider(),
			providers.NewBlocketProvider(),
			providers.NewKronofogdenProvider(),
		},
	)

	http.HandleFunc("GET /atom.xml", func(w http.ResponseWriter, req *http.Request) {
		logRequest(req)
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
	})

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		slog.Error("Error when parsing template", "error", err)
	}

	http.HandleFunc("GET /", func(w http.ResponseWriter, req *http.Request) {
		logRequest(req)
		queries, err := generator.GetQueries()
		if err != nil {
			slog.Error("Error when fetching queries", "error", err)
		}
		tmpl.Execute(w, queries)
	})

	http.HandleFunc("POST /queries/add", func(w http.ResponseWriter, req *http.Request) {
		logRequest(req)

		_ = req.ParseForm()
		query := req.Form.Get("query")
		if query == "" {
			w.WriteHeader(400)
		} else {
			queries, err := generator.AddQuery(query)
			if err != nil {
				slog.Error("Error creating query", "error", err)
				w.WriteHeader(400)
			} else {
				tmpl.Execute(w, queries)
			}
		}
	})

	http.HandleFunc("DELETE /queries/{id}", func(w http.ResponseWriter, req *http.Request) {
		logRequest(req)

		id := req.PathValue("id")
		queries, err := generator.DeleteQuery(id)
		if err != nil {
			slog.Error("Error deleting query", "error", err)
			w.WriteHeader(400)
		} else {
			tmpl.Execute(w, queries)
		}
	})

	http.ListenAndServe(":8080", nil)
}

func logRequest(req *http.Request) {
	dump, err := httputil.DumpRequest(req, false)
	if err != nil {
		slog.Error("Error dumping request", "error", err)
	} else {
		slog.Info("Got request", "request", dump)
	}
}
