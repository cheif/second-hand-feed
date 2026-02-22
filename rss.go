package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
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
	Queries []feedQuery `json:"queries"`
}

type feedQuery struct {
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

func (f *FeedGenerator) GetFeed() ([]byte, error) {
	config, err := f.getConfig()
	if err != nil {
		return nil, err
	}
	var items []Item
	for _, provider := range f.Providers {
		urls := config.getURLS(provider)
		providerItems, err := provider.GetItems(urls)
		if err != nil {
			log.Printf("Error when fetching items: %v", err)
		} else {
			items = append(items, providerItems...)
		}
	}

	var rssItems []RSSItem
	for _, item := range items {
		rssItems = append(rssItems, RSSItem{
			Link:  item.URL,
			Guid:  item.URL,
			Title: item.Title,
			Description: RSSDescription{
				Description: fmt.Sprintf("%v %v", item.Price.Amount, item.Price.CurrencyCode),
				CData:       fmt.Sprintf(`<img src="%v" />`, item.ImageURL),
			},
			PubDate: item.Timestamp.Format(time.RFC1123),
		})
	}

	rss := rss{
		Version: "2.0",
		Channel: channel{
			Title:       "Second hand rss",
			Link:        "",
			Description: "",
			Items:       rssItems,
		},
	}

	bytes, err := xml.MarshalIndent(rss, "", "  ")
	if err != nil {
		return nil, err
	}
	log.Printf("returning feed with %v items", len(items))
	return bytes, nil
}

type rss struct {
	Version string  `xml:"version,attr"`
	Channel channel `xml:"channel"`
}

type channel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Items       []RSSItem `xml:"item"`
}

type RSSItem struct {
	Link        string         `xml:"link"`
	Guid        string         `xml:"guid"`
	Title       string         `xml:"title"`
	Description RSSDescription `xml:"description"`
	PubDate     string         `xml:"pubDate"`
	Enclosure   RSSEnclosure   `xml:"enclosure"`
}

type RSSDescription struct {
	XMLName     xml.Name `xml:"description"`
	Description string   `xml:",innerxml"`
	CData       string   `xml:",cdata"`
}

type RSSEnclosure struct {
	URL string `xml:"url,attr"`
}
