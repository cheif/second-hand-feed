package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

type VintedProvider struct {
	client http.Client
}

func (f *VintedProvider) Name() string {
	return "vinted"
}

func NewVintedProvider() *VintedProvider {
	client := http.Client{}
	jar, _ := cookiejar.New(nil)
	client.Jar = jar
	return &VintedProvider{
		client: client,
	}
}

func (f *VintedProvider) GetItems(urls []url.URL) ([]Item, error) {
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
			slog.Error("Error when getting items", "url", url, "error", err)
		} else {
			items = append(items, queryItems...)
		}
	}
	return items, nil
}

func (f *VintedProvider) getItems(query url.URL) ([]Item, error) {
	url := getApiUrl(&query)
	slog.Info("Fetching vinted items", "url", url)
	resp, err := f.client.Get(url.String())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Unexpected status code: %v", resp.StatusCode)
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
	authURL := *query
	authURL.Path = ""
	authURL.RawQuery = ""
	_, err := f.client.Head(authURL.String())
	return err
}

func getApiUrl(query *url.URL) *url.URL {
	apiURL := *query
	apiURL.Path = "/api/v2/catalog/items"
	return &apiURL
}
