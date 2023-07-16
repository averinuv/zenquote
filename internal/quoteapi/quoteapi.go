package quoteapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const (
	apiURL        = "http://zenquotes.io"
	apiRandomPath = "/api/random"
)

var ErrEmptyQuotes = errors.New("received empty quotes")

type Quote struct {
	Quote  string `json:"q"`
	Author string `json:"a"`
}

type QuoteAPI struct {
	httpClient *http.Client
}

func NewQuoteAPI(httpClient *http.Client) *QuoteAPI {
	return &QuoteAPI{
		httpClient: httpClient,
	}
}

// GetRandom returns a random Zen quote or an error if one occurs.
func (q *QuoteAPI) GetRandom(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s%s", apiURL, apiRandomPath)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("new request failed: %w", err)
	}

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch quote: %w", err)
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var quote []Quote
	if err = json.Unmarshal(body, &quote); err != nil {
		return "", fmt.Errorf("failed to parse response body: %w", err)
	}

	if len(quote) == 0 {
		return "", ErrEmptyQuotes
	}

	return quote[0].Quote, nil
}
