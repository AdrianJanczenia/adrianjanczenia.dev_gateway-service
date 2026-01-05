package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/logic/errors"
)

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (c *Client) ProxyCVDownload(ctx context.Context, w http.ResponseWriter, token string, lang string) error {
	params := url.Values{}
	params.Add("token", token)
	params.Add("lang", lang)

	fullUrl := fmt.Sprintf("%s/download/cv?%s", c.baseURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, "GET", fullUrl, nil)
	if err != nil {
		return errors.ErrInternalServerError
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return errors.ErrServiceUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.FromHTTPStatus(resp.StatusCode)
	}

	allowList := map[string]bool{
		"Content-Type":        true,
		"Content-Disposition": true,
		"Content-Length":      true,
		"Last-Modified":       true,
	}

	for key, values := range resp.Header {
		if allowList[key] {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
	}

	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)

	return err
}
