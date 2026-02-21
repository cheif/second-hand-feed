package main

import (
	"encoding/xml"
	"fmt"
	"log"
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
	GetItems() ([]Item, error)
}

type FeedGenerator struct {
	Providers []ItemProvider
}

func (f *FeedGenerator) GetFeed() ([]byte, error) {
	var items []Item
	for _, provider := range f.Providers {
		providerItems, err := provider.GetItems()
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
