package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html"
	"log"
	"log/slog"
	"net/url"
	"os"
	"time"
)

type Item struct {
	URL       string
	Title     string
	Timestamp time.Time
	ImageURL  string
	Price     ItemPrice
}

type ItemPrice struct {
	Amount       string
	CurrencyCode string
}

type ItemProvider interface {
	Name() string
	GetItems(urls []url.URL) ([]Item, error)
	CanHandle(query url.URL) bool
}

type FeedGenerator struct {
	configPath string
	Providers  []ItemProvider
}

func NewFeedGenerator(configPath string, providers []ItemProvider) *FeedGenerator {
	return &FeedGenerator{
		configPath: configPath,
		Providers:  providers,
	}
}

type feedConfig struct {
	Queries []FeedQuery `json:"queries"`
}

type FeedQuery struct {
	Query    string `json:"query"`
	Provider string `json:"provider"`
}

func (f *FeedGenerator) getConfig() (*feedConfig, error) {
	data, err := os.ReadFile(f.configPath)
	if err != nil {
		slog.Error("Error when reading config", "error", err)
		return nil, err
	}
	var config feedConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (f *FeedGenerator) writeConfig(config feedConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(f.configPath, data, 0666)
	if err != nil {
		return err
	}
	return nil
}

func (c feedConfig) getURLS(provider ItemProvider) []url.URL {
	var urls []url.URL
	for _, query := range c.Queries {
		if query.Provider == provider.Name() {
			url, err := url.Parse(query.Query)
			if err != nil {
				slog.Error("Error parsing Query", "query", query.Query, "error", err)
			} else {
				urls = append(urls, *url)
			}
		}
	}
	return urls
}

func (f *FeedGenerator) GetQueries() ([]FeedQuery, error) {
	config, err := f.getConfig()
	if err != nil {
		return nil, err
	}
	return config.Queries, nil
}

func (f *FeedGenerator) AddQuery(query string) error {
	config, err := f.getConfig()
	if err != nil {
		return err
	}
	url, err := url.Parse(query)
	if err != nil {
		return err
	}
	for _, provider := range f.Providers {
		if provider.CanHandle(*url) {
			config.Queries = append(config.Queries, FeedQuery{
				Query:    query,
				Provider: provider.Name(),
			})
			err = f.writeConfig(*config)
			if err != nil {
				return err
			} else {
				return nil
			}
		}
	}
	return fmt.Errorf("No provider can handle: %v", query)
}

func (f *FeedGenerator) GetFeed(baseURL url.URL) ([]byte, error) {
	config, err := f.getConfig()
	if err != nil {
		return nil, err
	}
	var entries []atomEntry
	var lastUpdate time.Time
	for _, provider := range f.Providers {
		urls := config.getURLS(provider)
		providerItems, err := provider.GetItems(urls)
		if err != nil {
			log.Printf("Error when fetching items: %v", err)
		} else {
			for _, item := range providerItems {
				if item.Timestamp.After(lastUpdate) {
					lastUpdate = item.Timestamp
				}
				entries = append(entries, atomEntry{
					Id:    item.URL,
					Title: item.Title,
					Author: atomPerson{
						provider.Name(),
					},
					Link: atomLink{
						Href: item.URL,
					},
					Updated: item.Timestamp,
					Summary: atomText{
						Type:    "html",
						Content: html.EscapeString(fmt.Sprintf(`%v %v<br /><img src="%v" />`, item.Price.Amount, item.Price.CurrencyCode, item.ImageURL)),
					},
				})
			}
		}
	}

	// TODO: Use url as id
	feed := atomFeed{
		Namespace: "http://www.w3.org/2005/Atom",
		Id:        baseURL.String(),
		Link: atomLink{
			Href: baseURL.String(),
		},
		Title:   "Second hand",
		Updated: lastUpdate,
		Entry:   entries,
	}

	bytes, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return nil, err
	}
	slog.Info("Returning atom feed", "entries", len(entries))
	return bytes, nil
}

type atomFeed struct {
	XMLName   xml.Name    `xml:"feed"`
	Namespace string      `xml:"xmlns,attr"`
	Id        string      `xml:"id"`
	Link      atomLink    `xml:"link"`
	Title     string      `xml:"title"`
	Updated   time.Time   `xml:"updated"`
	Entry     []atomEntry `xml:"entry"`
}

type atomEntry struct {
	Id      string     `xml:"id"`
	Title   string     `xml:"title"`
	Updated time.Time  `xml:"updated"`
	Author  atomPerson `xml:"author"`
	Link    atomLink   `xml:"link"`
	Summary atomText   `xml:"summary"`
}

type atomPerson struct {
	Name string `xml:"name"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
}

type atomText struct {
	Type    string `xml:"type,attr"`
	Content string `xml:",innerxml"`
}
