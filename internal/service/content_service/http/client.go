package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/logic/errors"
)

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = 10 * time.Second

	return &Client{
		baseURL: baseURL,
		client:  &http.Client{Transport: transport},
	}
}

func (c *Client) ProxyCVDownload(ctx context.Context, w http.ResponseWriter, token string, lang string) error {
	params := url.Values{}
	params.Add("token", token)
	params.Add("lang", lang)

	fullURL := fmt.Sprintf("%s/download/cv?%s", c.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return errors.ErrInternalServerError
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return errors.ErrServiceUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, err := io.ReadAll(io.LimitReader(resp.Body, 4097))
		if err != nil || len(data) > 4096 {
			return errors.ErrBadGateway
		}

		var body struct {
			Error string `json:"error"`
		}

		if err := json.Unmarshal(data, &body); err != nil {
			return errors.ErrBadGateway
		}

		return errors.FromUpstream(resp.StatusCode, body.Error)
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
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")

	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)

	return err
}
