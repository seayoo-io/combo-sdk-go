package combo

import (
	"context"
)

type GetMiniGameWeixinAccessTokenInput struct {
	// 微信小游戏的 AppID。必须是当前游戏配置的微信小游戏应用。
	AppId string `json:"app_id"`
}

type GetMiniGameWeixinAccessTokenOutput struct {
	baseResponse

	// 微信小游戏的 AppID，与请求中的 AppId 一致。
	AppId string `json:"app_id"`

	// 微信小游戏的接口调用凭证。
	AccessToken string `json:"access_token"`
}

// 获取微信小游戏的接口调用凭证，供游戏服务端调用微信服务端 API 使用。
//
// 接口调用凭证由世游服务端统一维护和刷新，游戏侧无需自行调用微信接口获取。
//
// 此接口仅适用于微信小游戏。
//
// 游戏服务端应当缓存接口调用凭证，缓存时间 1 分钟，按照 1 次/分钟 的频率请求该 API 获取并刷新缓存。
func (c *Client) GetMiniGameWeixinAccessToken(ctx context.Context, input *GetMiniGameWeixinAccessTokenInput) (*GetMiniGameWeixinAccessTokenOutput, error) {
	output := &GetMiniGameWeixinAccessTokenOutput{}
	err := c.callApi(ctx, "minigame-weixin-access-token", input, output)
	if err != nil {
		return nil, err
	}
	return output, nil
}
