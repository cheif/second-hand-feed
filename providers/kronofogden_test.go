package providers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

func TestFetchAndParse(t *testing.T) {
	server := createTestServer(t)
	defer server.Close()
	provider := KronofogdenProvider{
		client: server.Client(),
	}

	testURL, _ := url.Parse(server.URL)
	testURL.Path = "/auk/w.objectlist"
	query := url.Values{}
	query.Add("inC", "KFM")
	query.Add("inA", "WEB")
	query.Add("inSite", "Y")
	testURL.RawQuery = query.Encode()

	items, err := provider.GetItems([]url.URL{*testURL})
	if err != nil {
		t.Error(err)
	}
	if len(items) != 7 {
		t.Errorf("Incorrect number of items returned: %v", len(items))
	}

	itemURL := testURL
	itemURL.Path = "/auk/w.object"
	query = url.Values{}
	query.Add("inC", "KFM")
	query.Add("inA", "20260218_1544")
	query.Add("inO", "1")
	itemURL.RawQuery = query.Encode()

	expected := Item{
		URL:   itemURL.String(),
		Title: "Grovdammsugare",
		//Timestamp time.Time
		ImageURL: "https://pic09.auction2000.online/aukpic/kfm/20260218_1544/111848_1_thumb.jpg?0644",
		Price: ItemPrice{
			Amount:       "1200",
			CurrencyCode: "SEK",
		},
	}
	if items[0] != expected {
		t.Errorf("Unexpected first \n    item: %v, \nexpected: %v", items[0], expected)
	}
}

func TestCanHandle(t *testing.T) {
	server := createTestServer(t)
	defer server.Close()
	provider := KronofogdenProvider{
		client: server.Client(),
	}

	testURL, _ := url.Parse(server.URL)
	testURL.Path = "/auk/w.objectlist?inC=KFM&inA=WEB&inSite=Y"

	query := provider.CanHandle(*testURL)

	expected := FeedQuery{
		Title:    "",
		Query:    testURL.String(),
		Provider: "kronofogden",
	}
	if *query != expected {
		t.Errorf("Unexpected handle \nresponse: %v, \nexpected: %v", *query, expected)
	}

}

func createTestServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		data, err := os.ReadFile("testdata/kronofogden.html")
		if err != nil {
			t.Error(err)
		}
		w.Write(data)
	}))
}
