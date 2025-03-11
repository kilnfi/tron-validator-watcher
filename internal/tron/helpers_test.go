package tron

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func setup(t *testing.T, mux *http.ServeMux) *Client {
	t.Helper()

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client, err := NewClient(
		WithBaseURL(server.URL),
	)
	if err != nil {
		t.Fatal(err)
	}
	return client
}
