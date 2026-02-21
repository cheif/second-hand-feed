package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"
)

type VintedProvider struct {
	client     http.Client
	configPath string
}

type VintedConfig struct {
	Urls []string `json:"urls"`
}

func NewVintedProvider(configPath string) *VintedProvider {
	client := http.Client{}
	jar, _ := cookiejar.New(nil)
	client.Jar = jar
	return &VintedProvider{
		client:     client,
		configPath: configPath,
	}
}

func (f *VintedProvider) GetURLs() []url.URL {
	data, err := os.ReadFile(f.configPath)
	if err != nil {
		log.Printf("Error reading config: %v", err)
		return nil
	}
	var config VintedConfig
	err = json.Unmarshal(data, &config)
	var urls []url.URL
	for _, rawUrl := range config.Urls {
		url, err := url.Parse(rawUrl)
		if err != nil {
			log.Printf("Error parsing url: %v err: %v", rawUrl, err)
		} else {
			urls = append(urls, *url)
		}

	}
	return urls
}

func (f *VintedProvider) GetItems() ([]Item, error) {
	urls := f.GetURLs()
	if len(urls) == 0 {
		return nil, nil
	}
	err := f.authenticate(&urls[0])
	if err != nil {
		return nil, err
	}
	var items []Item
	for _, url := range urls {
		queryItems, err := f.getItems(url)
		if err != nil {
			log.Printf("Error when getting items from URL: %v, err: %v", url, err)
		} else {
			items = append(items, queryItems...)
		}
	}
	return items, nil
}

func (f *VintedProvider) getItems(query url.URL) ([]Item, error) {
	url := getApiUrl(&query)
	resp, err := f.client.Get(url.String())
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var response vintedItemsResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	var items []Item
	for _, item := range response.Items {
		items = append(items, Item{
			URL:       item.URL,
			Title:     item.Title,
			Timestamp: time.Unix(int64(item.Photo.HighResolution.Timestamp), 0),
			ImageURL:  item.Photo.FullSizeURL,
			Price: ItemPrice{
				Amount:       item.TotalItemPrice.Amount,
				CurrencyCode: item.TotalItemPrice.CurrencyCode,
			},
		})
	}
	return items, nil
}

type vintedItemsResponse struct {
	Items []vintedItemResponse `json:"items"`
}

type vintedItemResponse struct {
	Title          string          `json:"title"`
	URL            string          `json:"url"`
	Photo          vintedItemPhoto `json:"photo"`
	TotalItemPrice vintedItemPrice `json:"total_item_price"`
}

type vintedItemPhoto struct {
	FullSizeURL    string                        `json:"full_size_url"`
	HighResolution vintedItemPhotoHighResolution `json:"high_resolution"`
}

type vintedItemPhotoHighResolution struct {
	Timestamp int `json:"timestamp"`
}

type vintedItemPrice struct {
	Amount       string `json:"amount"`
	CurrencyCode string `json:"currency_code"`
}

func (f *VintedProvider) authenticate(query *url.URL) error {
	authURL := query
	authURL.Path = ""
	authURL.RawQuery = ""
	_, err := f.client.Head(authURL.String())
	return err
}

func getApiUrl(query *url.URL) *url.URL {
	query.Path = "/api/v2/catalog/items"
	return query
}
