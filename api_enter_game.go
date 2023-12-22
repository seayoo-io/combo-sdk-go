package combo

import (
	"context"
)

type EnterGameInput struct {
	// 聚合用户标识。
	ComboId string `json:"combo_id"`

	// 游戏会话标识。
	// 单次游戏会话的上下线动作必须使用同一会话标识上报。
	SessionId string `json:"session_id"`
}

type EnterGameOutput struct {
	baseResponse

	// 暂时没有返回值。
}

func (c *client) EnterGame(ctx context.Context, input *EnterGameInput) (*EnterGameOutput, error) {
	output := &EnterGameOutput{}
	err := c.callApi(ctx, "enter-game", input, output)
	if err != nil {
		return nil, err
	}
	return output, nil
}
