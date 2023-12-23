package combo

import (
	"context"
)

type CreateOrderInput struct {
	// 用于标识创建订单请求的唯一 ID。
	ReferenceId string `json:"reference_id"`

	// 发起购买的用户的唯一标识。
	ComboId string `json:"combo_id"`

	// 要购买的商品 ID。
	ProductId string `json:"product_id"`

	// 客户端的运行平台。
	Platform Platform `json:"platform"`

	// 游戏侧异步接收发货通知的地址。
	NotifyUrl string `json:"notify_url"`

	// 要购买的商品的数量。
	Quantity int `json:"quantity,omitempty"`

	// 订单上下文，在发货通知中透传回游戏。
	Context string `json:"context,omitempty"`

	// 订单的元数据。
	Meta OrderMeta `json:"meta,omitempty"`
}

// OrderMeta 包含了订单的元数据。
//
// 大部分元数据用于数据分析与查询，游戏侧应当尽量提供。
//
// 某些元数据在特定的支付场景下是必须的，例如微信小游戏的 iOS 支付场景。
type OrderMeta struct {
	// 游戏大区 ID。
	ZoneId string `json:"zone_id,omitempty"`

	// 游戏服务器 ID。
	ServerId string `json:"server_id,omitempty"`

	// 游戏角色 ID。
	RoleId string `json:"role_id,omitempty"`

	// 游戏角色名。
	RoleName string `json:"role_name,omitempty"`

	// 游戏角色的等级。
	RoleLevel int `json:"role_level,omitempty"`

	// 微信小游戏的 App ID。
	// 微信小游戏的 iOS 支付场景必须传入，即 Platform == Platform_Weixin
	WeixinAppid string `json:"weixin_appid,omitempty"`

	// 微信小游戏的玩家 OpenID。
	// 微信小游戏的 iOS 支付场景必须传入，即 Platform == Platform_Weixin
	WeixinOpenid string `json:"weixin_openid,omitempty"`
}

type CreateOrderOutput struct {
	baseResponse

	// 世游服务端创建的，标识订单的唯一 ID。
	OrderId string `json:"order_id"`

	// 世游服务端创建的订单 token，用于后续支付流程。
	OrderToken string `json:"order_token"`

	// 订单失效时间。Unix timestamp in seconds。
	ExpiresAt int64 `json:"expires_at"`
}

// 创建订单，发起一个应用内购买 + 支付的流程。
func (c *Client) CreateOrder(ctx context.Context, input *CreateOrderInput) (*CreateOrderOutput, error) {
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
