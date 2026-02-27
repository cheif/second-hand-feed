package providers

import (
	"crypto/sha256"
	"fmt"
	"net/url"
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
	CanHandle(query url.URL) *FeedQuery
}

type FeedQuery struct {
	Title    string `json:"title"`
	Query    string `json:"query"`
	Provider string `json:"provider"`
}

func (q FeedQuery) Id() string {
	sum := sha256.Sum256([]byte(q.Query))
	return fmt.Sprintf("%x", sum)
}
