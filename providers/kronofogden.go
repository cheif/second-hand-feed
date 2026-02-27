package providers

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strings"

	"golang.org/x/net/html"
)

type KronofogdenProvider struct {
	client *http.Client
}

func NewKronofogdenProvider() *KronofogdenProvider {
	return &KronofogdenProvider{
		client: http.DefaultClient,
	}
}

func (k *KronofogdenProvider) Name() string {
	return "kronofogden"
}

func (k *KronofogdenProvider) CanHandle(query url.URL) *FeedQuery {
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

func (k *KronofogdenProvider) GetItems(urls []url.URL) ([]Item, error) {
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

func (k *KronofogdenProvider) fetch(url url.URL) (*kronofogdenResponse, error) {
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
	var response kronofogdenResponse
	for node := range doc.Descendants() {
		if isItemNode(node) {
			item, err := parseItemNode(node, url)
			if err != nil {
				slog.Error("Error parsing node", "node", node, "error", err)
			} else {
				response.items = append(response.items, *item)
			}
		}
	}
	return &response, nil
}

func isItemNode(node *html.Node) bool {
	return node.Type == html.ElementNode && node.Data == "div" && hasClass(node, "obj_thumbnail")
}

func parseItemNode(node *html.Node, baseURL url.URL) (*Item, error) {
	var err error
	var item Item
	for child := range node.Descendants() {
		if hasClass(child, "obj_txt_inner") {
			item.Title, err = parseTitle(child)
			if err != nil {
				return nil, err
			}
		} else if hasClass(child, "obj_link") {
			item.URL, err = parseURL(child, &baseURL)
			if err != nil {
				return nil, err
			}
		} else if hasClass(child, "obj_img") {
			item.ImageURL, err = getAttr(child, "src")
			if err != nil {
				return nil, err
			}
		} else if hasClass(child, "nico27") {
			price, err := parsePrice(child)
			if err == nil {
				item.Price = *price
			}
		}
	}
	return &item, nil
}

func parseTitle(node *html.Node) (string, error) {
	for child := range node.ChildNodes() {
		if child.Type == html.TextNode {
			trimmedContents := strings.TrimSpace(child.Data)
			if len(trimmedContents) > 0 {
				return trimmedContents, nil
			}
		}
	}
	return "", fmt.Errorf("No title found in node: %v", node)
}

func parseURL(node *html.Node, baseURL *url.URL) (string, error) {
	itemPath, err := getAttr(node, "href")
	if err != nil {
		return "", err
	}

	parsed, err := url.Parse(itemPath)
	if err != nil {
		return "", err
	}
	url := *baseURL
	url.RawQuery = parsed.Query().Encode()
	url.Path = path.Join(path.Dir(baseURL.Path), parsed.Path)
	return url.String(), nil
}

func parsePrice(node *html.Node) (*ItemPrice, error) {
	for child := range node.ChildNodes() {
		if child.Type == html.TextNode {
			trimmedContents := strings.TrimSpace(child.Data)
			splits := strings.SplitN(trimmedContents, " ", 2)
			if len(splits) == 2 {
				price := ItemPrice{
					Amount:       splits[0],
					CurrencyCode: splits[1],
				}
				return &price, nil
			}
		}
	}
	return nil, fmt.Errorf("No title found in node: %v", node)
}

func getAttr(node *html.Node, key string) (string, error) {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val, nil
		}
	}
	return "", fmt.Errorf("Could not find attr: %v in node: %v", key, node)
}

func hasClass(node *html.Node, class string) bool {
	for _, attr := range node.Attr {
		if attr.Key == "class" && strings.Contains(attr.Val, class) {
			return true
		}
	}
	return false
}

type kronofogdenResponse struct {
	items []Item
}
