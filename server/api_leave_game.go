package server

import (
	"context"

	"github.com/seayoo-io/combo-sdk-go/combo"
)

type LeaveGameInput struct {
	// 聚合用户标识。
	ComboId string `json:"combo_id"`

	// 游戏会话标识。
	// 单次游戏会话的上下线动作必须使用同一会话标识上报。
	SessionId string `json:"session_id"`
}

type LeaveGameOutput struct {
	combo.BaseResponse

	// 暂时没有返回值。
}

func (c *client) LeaveGame(ctx context.Context, input *LeaveGameInput) (*LeaveGameOutput, error) {
	output := &LeaveGameOutput{}
	err := c.callApi(ctx, "leave-game", input, output)
	if err != nil {
		return nil, err
	}
	return output, nil
}
