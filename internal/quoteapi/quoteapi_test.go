package quoteapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"zenquote/internal/quoteapi"

	"github.com/stretchr/testify/assert"
)

func TestGetRandom(t *testing.T) {
	t.Parallel()

	t.Run("successful response", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, r *http.Request) {
			jsonResponse := `[{"q": "Some random Zen quote", "a": "Unknown"}]`
			_, _ = w.Write([]byte(jsonResponse))
		}
		server := httptest.NewTLSServer(http.HandlerFunc(handler))
		defer server.Close()

		httpClient := server.Client()
		httpClient.Transport = &http.Transport{
			TLSClientConfig: server.Client().Transport.(*http.Transport).TLSClientConfig,
			Proxy: func(_ *http.Request) (*url.URL, error) {
				return url.Parse(server.URL)
			},
		}

		quoteAPI := quoteapi.NewQuoteAPI(httpClient)
		quote, err := quoteAPI.GetRandom(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, "Some random Zen quote", quote)
	})

	t.Run("no quotes in response", func(t *testing.T) {
		t.Parallel()

		handler := func(w http.ResponseWriter, r *http.Request) {
			jsonResponse := `[]` // Ответ без цитат
			_, _ = w.Write([]byte(jsonResponse))
		}
		server := httptest.NewTLSServer(http.HandlerFunc(handler))
		defer server.Close()

		httpClient := server.Client()
		httpClient.Transport = &http.Transport{
			TLSClientConfig: server.Client().Transport.(*http.Transport).TLSClientConfig,
			Proxy: func(_ *http.Request) (*url.URL, error) {
				return url.Parse(server.URL)
			},
		}

		quoteAPI := quoteapi.NewQuoteAPI(httpClient)
		quote, err := quoteAPI.GetRandom(context.Background())

		assert.EqualError(t, err, "received empty quotes")
		assert.Empty(t, quote)
	})
}
