package providers

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

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
