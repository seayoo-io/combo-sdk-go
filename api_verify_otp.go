package combo

import (
	"context"
)

type VerifyOtpInput struct {
	// 要验证验证码的用户的唯一标识，需与发送时一致。
	ComboId string `json:"combo_id"`

	// 发送验证码的通道，需与发送时一致。
	//
	// 不填写时默认为 sms。
	Channel OtpChannel `json:"channel"`

	// 发送验证码的目标行为，需与发送时一致。
	Action string `json:"action"`

	// 用户输入的验证码。
	Otp string `json:"otp"`
}

type VerifyOtpOutput struct {
	baseResponse

	// 是否验证通过。
	//
	// true 表示验证通过。验证通过后验证码立即失效，不可重复使用。
	// false 表示验证失败，用户输入的验证码与系统生成的验证码不匹配，
	// 可能是由于输入错误或验证码已过期，建议用户重新检查输入的验证码。
	Valid bool `json:"valid"`
}

// 验证世游通行证验证码。
//
// 验证 SendOtp 发送的验证码是否正确。验证成功后验证码立即失效，不可重复使用。
func (c *Client) VerifyOtp(ctx context.Context, input *VerifyOtpInput) (*VerifyOtpOutput, error) {
	if input.Channel == "" {
		input.Channel = OtpChannel_SMS
	}
	output := &VerifyOtpOutput{}
	err := c.callApi(ctx, "verify-otp", input, output)
	if err != nil {
		return nil, err
	}
	return output, nil
}
