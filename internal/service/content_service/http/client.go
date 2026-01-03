package http

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
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

func (c *Client) ProxyCVDownload(w http.ResponseWriter, token string, lang string) error {
	params := url.Values{}
	params.Add("token", token)
	params.Add("lang", lang)

	fullUrl := fmt.Sprintf("%s/download/cv?%s", c.baseURL, params.Encode())
	req, err := http.NewRequest("GET", fullUrl, nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)

	if resp.StatusCode == http.StatusOK {
		_, err = io.Copy(w, resp.Body)
		return err
	}

	return nil
}
