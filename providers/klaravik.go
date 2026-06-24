package providers

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type KlaravikProvider struct {
	client *http.Client
}

func NewKlaravikProvider() *KlaravikProvider {
	return &KlaravikProvider{
		client: http.DefaultClient,
	}
}

func (k *KlaravikProvider) Name() string {
	return "klaravik"
}

func (k *KlaravikProvider) CanHandle(query url.URL) *FeedQuery {
	if !strings.Contains(query.Host, "klaravik") {
		return nil
	}
	_, err := k.fetch(query)
	if err != nil {
		return nil
	}
	return &FeedQuery{
		Title:    "",
		Query:    query.String(),
		Provider: k.Name(),
	}
}

func (k *KlaravikProvider) GetItems(urls []url.URL) ([]Item, error) {
	var items []Item
	for _, url := range urls {
		resp, err := k.fetch(urls[0])

		if err != nil {
			slog.Error("Error when fetching items for url", "url", url.String(), "error", err)
		} else {
			items = append(items, resp.items...)
		}

	}
	return items, nil
}

func (k *KlaravikProvider) fetch(url url.URL) (*klaravikResponse, error) {
	resp, err := k.client.Get(url.String())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Unexpected status code: %v", resp.StatusCode)
	}
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}
	var response klaravikResponse
	for node := range doc.Descendants() {
		if isListingNode(node) {
			item, err := parseListingNode(node, url)
			if err != nil {
				slog.Error("Error parsing node", "node", node, "error", err)
			} else {
				response.items = append(response.items, *item)
			}
		}
	}
	return &response, nil
}

func isListingNode(node *html.Node) bool {
	return node.Type == html.ElementNode && node.Data == "div" && hasClass(node, "isListing")
}

func parseListingNode(node *html.Node, baseURL url.URL) (*Item, error) {
	var err error
	var item Item
	for child := range node.Descendants() {
		if child.Type == html.ElementNode && child.Data == "a" {
			item.Title, err = getAttr(child, "title")
			if err != nil {
				return nil, err
			}
			itemURL, err := parseKlaravikURL(child, &baseURL)
			if err != nil {
				return nil, err
			}
			item.URL = itemURL.String()
		} else if child.Type == html.ElementNode && child.Data == "img" && item.ImageURL == "" {
			item.ImageURL, err = getAttr(child, "src")
			if err != nil {
				return nil, err
			}
		} else if hasClass(child, "product_card__mark-fav") {
			timestampStr, err := getAttr(child, "data-auction-start")
			if err != nil {
				return nil, err
			}
			item.Timestamp, err = time.Parse(time.RFC3339, timestampStr)
			if err != nil {
				return nil, err
			}
		} else if hasClass(child, "product_card__current-bid") {
			price, err := parseKlaravikPrice(child)
			if err == nil {
				item.Price = *price
			}
		} else if hasClass(child, "product_card__info-text") {
			item.Location = &ItemLocation{
				Name: child.FirstChild.Data,
			}
		}
	}
	return &item, nil
}

func parseKlaravikURL(node *html.Node, baseURL *url.URL) (*url.URL, error) {
	itemPath, err := getAttr(node, "href")
	if err != nil {
		return nil, err
	}

	parsed, err := url.Parse(itemPath)
	if err != nil {
		return nil, err
	}
	url := *baseURL
	url.RawQuery = parsed.Query().Encode()
	url.Path = parsed.Path
	return &url, nil
}

func parseKlaravikPrice(node *html.Node) (*ItemPrice, error) {
	child := node.FirstChild
	trimmedContents := strings.TrimSpace(child.Data)
	splits := strings.Fields(trimmedContents)
	if len(splits) >= 2 {
		price := ItemPrice{
			Amount:       strings.Join(splits[:len(splits)-1], ""),
			CurrencyCode: splits[len(splits)-1],
		}
		return &price, nil
	}
	return nil, fmt.Errorf("No price found in node: %v", node)
}

type klaravikResponse struct {
	items []Item
}
