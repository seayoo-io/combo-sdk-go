package combo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client 是一个用来调用 Combo Server API 的 API Client
type Client struct {
	endpoint  Endpoint
	client    HttpClient
	signer    httpSigner
	userAgent string
}

// ClientOption 是函数式风格的的可选项，用于创建 Client。
type ClientOption func(*Client)

// HttpClient 用于发送 HTTP 请求。如果需要对 HTTP 请求的行为和参数进行自定义设置，可以实现此接口。
//
// 通常来说 *http.Client 可以满足绝大部分需求。
type HttpClient interface {
	Do(*http.Request) (*http.Response, error)
}

// WithHttpClient 用于指定自定义的 HttpClient。
//
// 如果不指定 HttpClient，则默认使用 http.DefaultClient。
func WithHttpClient(client HttpClient) ClientOption {
	return func(c *Client) {
		c.client = client
	}
}

// NewClient 创建一个新的 Server API 的 client
func NewClient(cfg Config, options ...ClientOption) (*Client, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	c := &Client{
		endpoint: cfg.Endpoint,
		signer: httpSigner{
			game:       cfg.GameId,
			signingKey: cfg.SecretKey,
		},
		userAgent: userAgent(cfg.GameId),
	}
	for _, option := range options {
		option(c)
	}
	if c.client == nil {
		c.client = http.DefaultClient
	}
	return c, nil
}

func (c *Client) callApi(ctx context.Context, api string, input any, output responseReader) error {
	req, err := c.newHttpRequest(ctx, api, input)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorResponse := &ErrorResponse{}
		if err := errorResponse.readResponse(resp); err != nil {
			return fmt.Errorf("error reading error response: %w", err)
		}
		return errorResponse
	}

	if err := output.readResponse(resp); err != nil {
		return fmt.Errorf("error reading response: %w", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}
	if err := json.Unmarshal(body, output); err != nil {
		return fmt.Errorf("failed to unmarshal response body: %w", err)
	}
	return nil
}

func (c *Client) newHttpRequest(ctx context.Context, api string, input any) (*http.Request, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	url := c.endpoint.url(api)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	c.setRequestHeaders(req)
	err = c.signRequest(req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (c *Client) setRequestHeaders(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Content-Type", "application/json")
}

func (c *Client) signRequest(req *http.Request) error {
	return c.signer.SignHttp(req, time.Now())
}
