package combo

import (
	"context"
)

type VoiceModerationRequestInput struct {
	// 房间实例 ID，唯一标识某个语音房间的一次存续（从开启到关闭）。
	RoomInstanceId string `json:"room_instance_id"`

	// 游戏服务器 ID。
	ServerId uint32 `json:"server_id"`

	// 提交审核申请的玩家角色 ID。
	RequesterRoleId string `json:"requester_role_id"`

	// 提交审核申请的玩家的唯一标识。
	RequesterComboId string `json:"requester_combo_id"`

	// 被提交语音审核的玩家角色 ID 列表，一次最多提交 32 个。
	TargetRoleIds []string `json:"target_role_ids"`

	// 提交审核申请的原因列表，选填。
	//
	// 一次最多提交 12 个，单个原因最多 32 个字符，取值参见：审核申请原因维度表。
	// 注意：取值不能包含英文逗号。
	Reasons []string `json:"reasons,omitempty"`
}

type VoiceModerationRequestOutput struct {
	baseResponse

	// 暂时没有返回值。
}

// 申请语音审核。
//
// 玩家认为语音房间内某些玩家存在语音违规行为时，可提交语音审核申请。
func (c *Client) VoiceModerationRequest(ctx context.Context, input *VoiceModerationRequestInput) (*VoiceModerationRequestOutput, error) {
	output := &VoiceModerationRequestOutput{}
	err := c.callApi(ctx, "voice-moderation-request", input, output)
	if err != nil {
		return nil, err
	}
	return output, nil
}
