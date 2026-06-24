package providers

import (
	"net/url"
	"reflect"
	"testing"
	"time"
)

func TestBlocketFetchAndParse(t *testing.T) {
	server := createTestServer(t, "testdata/blocket.json")
	defer server.Close()
	provider := BlocketProvider{
		client: server.Client(),
	}

	testURL, _ := url.Parse(server.URL)

	items, err := provider.GetItems([]url.URL{*testURL})
	if err != nil {
		t.Error(err)
	}
	if len(items) != 3 {
		t.Errorf("Incorrect number of items returned: %v", len(items))
	}

	time, err := time.Parse(time.RFC3339, "2026-01-31T13:51:47+01:00")
	if err != nil {
		t.Error(err)
	}
	expected := Item{
		URL:       "https://www.blocket.se/recommerce/forsale/item/20523663",
		Title:     "Hemmagym weider",
		Timestamp: time,
		ImageURL:  "https://images.blocketcdn.se/dynamic/default/item/20523663/5169a664-2be0-48b3-8d41-a67c3c90974b",
		Price: ItemPrice{
			Amount:       "8000",
			CurrencyCode: "SEK",
		},
		Location: &ItemLocation{
			Name: "Ankarsvik",
		},
	}
	if !reflect.DeepEqual(items[0], expected) {
		t.Errorf("Unexpected first \n    item: %v, \nexpected: %v", items[0], expected)
	}
}

func TestBlocketCanHandle(t *testing.T) {
	server := createTestServer(t, "testdata/blocket.json")
	defer server.Close()
	provider := BlocketProvider{
		client: server.Client(),
	}

	testURL, _ := url.Parse("https://www.blocket.se/recommerce/forsale/search?location=0.300022&location=0.300024&q=prusa")

	query := provider.CanHandle(*testURL)

	expected := FeedQuery{
		Title:    "prusa, Västerbotten med flera",
		Query:    testURL.String(),
		Provider: "blocket",
	}
	if *query != expected {
		t.Errorf("Unexpected handle \nresponse: %v, \nexpected: %v", *query, expected)
	}

}
