package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/seayoo-io/combo-sdk-go/combo"
)

// Client 是一个用来调用 Combo Server API 的 API client
type Client interface {
	// 创建订单，发起一个应用内购买 + 支付的流程。
	CreateOrder(context.Context, *CreateOrderInput) (*CreateOrderOutput, error)

	// 通知世游服务端玩家进入游戏世界（上线）。
	//
	// 此接口仅用于中宣部防沉迷系统的上下线数据上报。
	EnterGame(context.Context, *EnterGameInput) (*EnterGameOutput, error)

	// 通知世游服务端玩家离开游戏世界（下线）。
	//
	// 此接口仅用于中宣部防沉迷系统的上下线数据上报。
	LeaveGame(context.Context, *LeaveGameInput) (*LeaveGameOutput, error)
}

// NewClient 创建一个新的 Server API 的 client
func NewClient(o Options) (Client, error) {
	if o.Endpoint.Url == "" {
		return nil, errors.New("missing required Endpoint")
	}
	o.Endpoint.Url = strings.TrimSuffix(o.Endpoint.Url, "/")
	if o.GameId.Id == "" {
		return nil, errors.New("missing required GameId")
	}
	if o.SecretKey.Key == nil || len(o.SecretKey.Key) == 0 {
		return nil, errors.New("missing required SecretKey")
	}
	if !strings.HasPrefix(string(o.SecretKey.Key), "sk_") {
		return nil, errors.New("invalid SecretKey: must start with sk_")
	}
	if o.HttpClient == nil {
		o.HttpClient = &http.Client{}
	}
	if o.Signer == nil {
		signer, err := NewHttpSigner(o.GameId, o.SecretKey)
		if err != nil {
			return nil, err
		}
		o.Signer = signer
	}
	return &client{
		options: o,
	}, nil
}

type client struct {
	options Options
}

func (c *client) callApi(ctx context.Context, api string, params any, result combo.ComboResponse) error {
	req, err := c.newHttpRequest(ctx, api, params)
	if err != nil {
		return err
	}
	resp, err := c.options.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorResponse := &combo.ErrorResponse{}
		if err := errorResponse.ReadHttp(resp); err != nil {
			return fmt.Errorf("error reading error response: %w", err)
		}
		return errorResponse
	}

	result.ReadHttp(resp)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}
	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to unmarshal response body: %w", err)
	}
	return nil
}

func (c *client) newHttpRequest(ctx context.Context, api string, params any) (*http.Request, error) {
	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	url := c.options.Endpoint.ServerApiUrl(api)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	setRequestHeaders(req)
	err = c.signRequest(req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (c *client) signRequest(req *http.Request) error {
	return c.options.Signer.SignHttp(req, time.Now())
}

func setRequestHeaders(req *http.Request) {
	// TODO: set User-Agent
	// req.Header.Set("User-Agent", "")
	req.Header.Set("Content-Type", "application/json")
}