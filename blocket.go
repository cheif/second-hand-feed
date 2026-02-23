package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type BlocketProvider struct {
	client http.Client
}

func NewBlocketProvider() *BlocketProvider {
	return &BlocketProvider{
		client: http.Client{},
	}
}

func (b *BlocketProvider) Name() string {
	return "blocket"
}

func (b *BlocketProvider) GetItems(urls []url.URL) ([]Item, error) {
	url, err := url.Parse("https://www.blocket.se/recommerce/forsale/search?location=0.300022&q=skivstång")
	if err != nil {
		return nil, err
	}
	return b.getItems(*url)
}

func (b *BlocketProvider) getItems(query url.URL) ([]Item, error) {
	url := getBlocketApiUrl(&query)
	slog.Info("Fetching blocket items", "url", url)
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:128.0) Gecko/20100101 Firefox/128.0")
	resp, err := b.client.Do(req)
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
	var response blocketResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	var items []Item
	for _, doc := range response.Docs {
		items = append(items, Item{
			URL:       doc.URL,
			Title:     doc.Heading,
			Timestamp: time.UnixMilli(int64(doc.Timestamp)),
			ImageURL:  doc.Image.URL,
			Price: ItemPrice{
				Amount:       strconv.Itoa(doc.Price.Amount),
				CurrencyCode: doc.Price.CurrencyCode,
			},
		})
	}

	return items, nil
}

type blocketResponse struct {
	Docs []blocketDoc `json:"docs"`
}

type blocketDoc struct {
	URL       string          `json:"canonical_url"`
	Heading   string          `json:"heading"`
	Timestamp int             `json:"timestamp"`
	Image     blocketDocImage `json:"image"`
	Price     blocketDocPrice `json:"price"`
}

type blocketDocImage struct {
	URL string `json:"url"`
}

type blocketDocPrice struct {
	Amount       int    `json:"amount"`
	CurrencyCode string `json:"currency_code"`
}

func getBlocketApiUrl(queryURL *url.URL) *url.URL {
	apiURL := *queryURL
	apiURL.Path += "/api/search/SEARCH_ID_BAP_COMMON"
	query := apiURL.Query()
	query.Add("sort", "PUBLISHED_DESC")
	apiURL.RawQuery = query.Encode()
	return &apiURL
}
