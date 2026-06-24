package providers

import (
	"net/url"
	"reflect"
	"testing"
	"time"
)

func TestKlaravikFetchAndParse(t *testing.T) {
	server := createTestServer(t, "testdata/klaravik.html")
	defer server.Close()
	provider := KlaravikProvider{
		client: server.Client(),
	}

	testURL, _ := url.Parse(server.URL)
	testURL.Path = "/auktion/"
	query := url.Values{}
	query.Add("searchtext", "Husqvarna")
	query.Add("setcountyflag", "265")
	testURL.RawQuery = query.Encode()

	items, err := provider.GetItems([]url.URL{*testURL})
	if err != nil {
		t.Error(err)
	}
	if len(items) != 1 {
		t.Errorf("Incorrect number of items returned: %v", len(items))
	}

	itemURL := testURL
	itemURL.Path = "/auktion/produkt/3235583-asfalltspadda-husqvarna-lf70/"
	itemURL.RawQuery = ""

	time, err := time.Parse(time.RFC3339, "2026-06-18T09:00:00+02:00")
	if err != nil {
		t.Error(err)
	}
	expected := Item{
		URL:       itemURL.String(),
		Title:     "Asfaltspadda Husqvarna LF70",
		Timestamp: time,
		ImageURL:  "https://media.se.klaravik.com/public/productimages/4f/8e/3235583/extrabilder73936430_thumblarge.jpg?v=1781682944",
		Price: ItemPrice{
			Amount:       "1500",
			CurrencyCode: "SEK",
		},
		Location: &ItemLocation{
			Name: "Timrå",
		},
	}
	if !reflect.DeepEqual(items[0], expected) {
		t.Errorf("Unexpected first \n    item: %v, \nexpected: %v", items[0], expected)
	}
}

func TestKlaravikCanHandle(t *testing.T) {
	server := createTestServer(t, "testdata/klaravik.html")
	defer server.Close()
	provider := KlaravikProvider{
		client: server.Client(),
	}

	testURL, _ := url.Parse("https://www.klaravik.se/auktion/?searchtext=Husqvarna&setcountyflag%5B%5D=265")

	query := provider.CanHandle(*testURL)

	expected := FeedQuery{
		Title:    "",
		Query:    testURL.String(),
		Provider: "klaravik",
	}
	if *query != expected {
		t.Errorf("Unexpected handle \nresponse: %v, \nexpected: %v", *query, expected)
	}

}
