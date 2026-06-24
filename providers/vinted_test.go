package providers

import (
	"net/url"
	"reflect"
	"testing"
	"time"
)

func TestVintedFetchAndParse(t *testing.T) {
	server := createTestServer(t, "testdata/vinted.json")
	defer server.Close()
	provider := VintedProvider{
		client: server.Client(),
	}

	testURL, _ := url.Parse(server.URL)

	items, err := provider.GetItems([]url.URL{*testURL})
	if err != nil {
		t.Error(err)
	}
	if len(items) != 48 {
		t.Errorf("Incorrect number of items returned: %v", len(items))
	}

	time, err := time.Parse(time.RFC3339, "2026-02-21T15:13:13+01:00")
	if err != nil {
		t.Error(err)
	}
	expected := Item{
		URL:       "https://www.vinted.se/items/8226425703-plecak-osprey-skarab-30-l",
		Title:     "Plecak Osprey Skarab 30 l",
		Timestamp: time,
		ImageURL:  "https://images1.vinted.net/tc/06_022d4_smpHzLSfMr9cHVr4JZKrz85k/1771683193.jpeg?s=deeeb31a0912682fb1c2116e918a5587e765d09c",
		Price: ItemPrice{
			Amount:       "809.99",
			CurrencyCode: "SEK",
		},
	}
	if !reflect.DeepEqual(items[0], expected) {
		t.Errorf("Unexpected first \n    item: %v, \nexpected: %v", items[0], expected)
	}
}

func TestVintedCanHandle(t *testing.T) {
	server := createTestServer(t, "testdata/vinted.json")
	defer server.Close()
	provider := VintedProvider{
		client: server.Client(),
	}

	testURL, _ := url.Parse("https://www.vinted.se/catalog?brand_ids[]=301297&search_text=archeon")

	query := provider.CanHandle(*testURL)

	expected := FeedQuery{
		Title:    "",
		Query:    testURL.String(),
		Provider: "vinted",
	}
	if *query != expected {
		t.Errorf("Unexpected handle \nresponse: %v, \nexpected: %v", *query, expected)
	}

}
