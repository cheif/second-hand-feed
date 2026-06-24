package providers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func createTestServer(t *testing.T, file string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Error(err)
		}
		w.Write(data)
	}))
}
