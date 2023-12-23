package combo

import (
	"context"
)

type LeaveGameInput struct {
	// 聚合用户标识。
	ComboId string `json:"combo_id"`

	// 游戏会话标识。
	// 单次游戏会话的上下线动作必须使用同一会话标识上报。
	SessionId string `json:"session_id"`
}

type LeaveGameOutput struct {
	baseResponse

	// 暂时没有返回值。
}

// 通知世游服务端玩家离开游戏世界（下线）。
//
// 此接口仅用于中宣部防沉迷系统的上下线数据上报。
func (c *Client) LeaveGame(ctx context.Context, input *LeaveGameInput) (*LeaveGameOutput, error) {
	output := &LeaveGameOutput{}
	err := c.callApi(ctx, "leave-game", input, output)
	if err != nil {
		return nil, err
	}
	return output, nil
}
