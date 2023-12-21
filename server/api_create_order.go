package server

import (
	"context"

	"github.com/seayoo-io/combo-sdk-go/combo"
)

type CreateOrderInput struct {
	// 用于标识创建订单请求的唯一 ID。
	ReferenceId string `json:"reference_id"`

	// 发起购买的用户的唯一标识。
	ComboId string `json:"combo_id"`

	// 要购买的商品 ID。
	ProductId string `json:"product_id"`

	// 客户端的运行平台。
	Platform combo.Platform `json:"platform"`

	// 游戏侧异步接收发货通知的地址。
	NotifyUrl string `json:"notify_url"`

	// 要购买的商品的数量。
	Quantity int `json:"quantity"`

	// 订单上下文，在发货通知中透传回游戏。
	Context string `json:"context"`

	// 订单的元数据。
	Meta OrderMeta `json:"meta"`
}

type OrderMeta struct {
	// 游戏大区 ID。
	// 用于数据分析与查询。
	ZoneId string `json:"zone_id"`

	// 游戏服务器 ID。
	// 用于数据分析与查询。
	ServerId string `json:"server_id"`

	// 游戏角色 ID。
	// 用于数据分析与查询。
	RoleId string `json:"role_id"`

	// 游戏角色名。
	// 用于数据分析与查询。
	RoleName string `json:"role_name"`

	// 游戏角色的等级。
	// 用于数据分析与查询。
	RoleLevel int `json:"role_level"`

	// 微信小游戏的 App ID。
	// 微信小游戏的 iOS 支付场景必须传入，即 Platform == combo.Platform_Weixin
	WeixinAppid string `json:"weixin_appid"`

	// 微信小游戏的玩家 OpenID。
	// 微信小游戏的 iOS 支付场景必须传入，即 Platform == combo.Platform_Weixin
	WeixinOpenid string `json:"weixin_openid"`
}

type CreateOrderOutput struct {
	combo.BaseResponse

	// 世游服务端创建的，标识订单的唯一 ID。
	OrderId string `json:"order_id"`

	// 世游服务端创建的订单 token，用于后续支付流程。
	OrderToken string `json:"order_token"`

	// 订单失效时间。Unix timestamp in seconds。
	ExpiresAt int64 `json:"expires_at"`
}

func (c *client) CreateOrder(ctx context.Context, input *CreateOrderInput) (*CreateOrderOutput, error) {
	if input.Quantity <= 0 {
		input.Quantity = 1
	}
	output := &CreateOrderOutput{}
	err := c.callApi(ctx, "create-order", input, output)
	if err != nil {
		return nil, err
	}
	return output, nil
}
