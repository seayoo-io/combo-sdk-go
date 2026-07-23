package combo

import (
	"context"
)

// OtpChannel 是发送验证码的通道。
type OtpChannel string

const (
	// 手机短信通道
	OtpChannel_SMS OtpChannel = "sms"
)

type SendOtpInput struct {
	// 要发送验证码的用户的唯一标识。
	ComboId string `json:"combo_id"`

	// 发送验证码的通道，目前仅支持 sms。
	//
	// 不填写时默认为 sms。
	Channel OtpChannel `json:"channel"`

	// 发送验证码的目标行为，由世游发行平台创建并管理。
	Action string `json:"action"`

	// 发送方元数据，主要用于数据分析，游戏服务端应当尽量提供。
	Meta OtpMeta `json:"meta,omitempty"`
}

// OtpMeta 包含了发送验证码的元数据。
//
// 元数据主要用于数据分析与查询，游戏侧应当尽量提供。
type OtpMeta struct {
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
}

type SendOtpOutput struct {
	baseResponse

	// 掩码后的手机号，仅当 channel=sms 时有值。
	Mobile string `json:"mobile"`

	// 验证码有效期，单位秒。
	OtpTtl int `json:"otp_ttl"`

	// 重新发送验证码的冷却时间，单位秒。
	OtpCooldown int `json:"otp_cooldown"`
}

// 发送世游通行证验证码。
//
// 根据 combo_id 向对应的世游通行证账号发送验证码，随后可调用 VerifyOtp 进行验证。
func (c *Client) SendOtp(ctx context.Context, input *SendOtpInput) (*SendOtpOutput, error) {
	if input.Channel == "" {
		input.Channel = OtpChannel_SMS
	}
	output := &SendOtpOutput{}
	err := c.callApi(ctx, "send-otp", input, output)
	if err != nil {
		return nil, err
	}
	return output, nil
}
