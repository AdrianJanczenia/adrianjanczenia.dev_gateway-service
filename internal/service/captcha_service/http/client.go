package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/logic/errors"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient(httpClient *http.Client, baseURL string) *Client {
	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

type GetPowResponse struct {
	Seed      string `json:"seed"`
	Signature string `json:"signature"`
}

func (c *Client) GetPow(ctx context.Context) (*GetPowResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/pow", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleError(resp)
	}

	var data GetPowResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}

type GetCaptchaRequest struct {
	Seed      string `json:"seed"`
	Signature string `json:"signature"`
	Nonce     string `json:"nonce"`
}

type GetCaptchaResponse struct {
	CaptchaId  string `json:"captchaId"`
	CaptchaImg string `json:"captchaImg"`
}

func (c *Client) GetCaptcha(ctx context.Context, body GetCaptchaRequest) (*GetCaptchaResponse, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/captcha", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleError(resp)
	}

	var data GetCaptchaResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}

type VerifyCaptchaRequest struct {
	CaptchaId    string `json:"captchaId"`
	CaptchaValue string `json:"captchaValue"`
}

type VerifyCaptchaResponse struct {
	CaptchaId string `json:"captchaId"`
}

func (c *Client) VerifyCaptcha(ctx context.Context, body VerifyCaptchaRequest) (*VerifyCaptchaResponse, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/verify", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleError(resp)
	}

	var data VerifyCaptchaResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}

func (c *Client) handleError(resp *http.Response) error {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.FromHTTPStatus(resp.StatusCode)
	}

	var body map[string]string
	if err := json.Unmarshal(bodyBytes, &body); err == nil {
		if slug, ok := body["error"]; ok {
			return errors.FromSlug(slug)
		}
	}

	return errors.FromHTTPStatus(resp.StatusCode)
}
