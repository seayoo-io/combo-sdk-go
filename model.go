package combo

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

// Combo API 端点。
type Endpoint string

// 由世游为游戏分配，用于标识游戏的业务代号。
type GameId string

// 由世游侧为游戏分配，游戏侧和世游侧共享的密钥。
// 此密钥用于签名计算与验证。
type SecretKey []byte

// 游戏客户端运行平台。
type Platform string

// Identity Provider (IdP) 是世游定义的用户身份提供方，俗称账号系统。
type IdP string

const (
	// 中国大陆 API 端点，用于国内发行
	Endpoint_China Endpoint = "https://api.seayoo.com"

	// 全球的 API 端点，用于海外发行
	Endpoint_Global Endpoint = "https://api.seayoo.io"
)

const (
	// 苹果的 iOS 和 iPadOS
	Platform_iOS Platform = "ios"

	// 安卓平台，包括华为鸿蒙系统、小米澎湃 OS 等基于 Android 的操作系统
	Platform_Android Platform = "android"

	// Windows (PC) 桌面平台
	Platform_Windows Platform = "windows"

	// macOS 桌面平台
	Platform_macOS Platform = "macos"

	// 微信小游戏
	Platform_Weixin Platform = "weixin"

	// 小游戏平台
	Platform_WebGL Platform = "webgl"
)

const (
	// 游客登录
	IdP_Guest IdP = "guest"

	// 世游通行证
	IdP_Seayoo IdP = "seayoo"

	// Sign in with Apple
	IdP_Apple IdP = "apple"

	// Google Account
	IdP_Google IdP = "google"

	// Facebook Login
	IdP_Facebook IdP = "facebook"

	// 小米账号
	IdP_Xiaomi IdP = "xiaomi"

	// 微信登录
	IdP_Weixin IdP = "weixin"

	// OPPO 账号
	IdP_Oppo IdP = "oppo"

	// VIVO 账号
	IdP_Vivo IdP = "vivo"

	// 华为账号
	IdP_Huawei IdP = "huawei"

	// 荣耀账号
	IdP_Honor IdP = "honor"

	// UC（九游）登录
	IdP_UC IdP = "uc"

	// TapTap 登录
	IdP_TapTap IdP = "taptap"

	// 哔哩哔哩（B站）账号
	IdP_Bilibili IdP = "bilibili"

	// 应用宝 YSDK 登录
	IdP_Yingyongbao IdP = "yingyongbao"

	// 4399 账号登录
	IdP_4399 IdP = "4399"

	// 抖音账号
	IdP_Douyin IdP = "douyin"

	// 雷电模拟器账号
	IdP_Leidian IdP = "leidian"

	// 猫窝游戏
	IdP_Maowo IdP = "maowo"

	// 联想
	IdP_Lenovo = "lenovo"

	// 魅族
	IdP_Meizu = "meizu"

	// 酷派
	IdP_Coolpad = "coolpad"

	// 努比亚
	IdP_Nubia = "nubia"
)

func (e Endpoint) url(api string) string {
	return fmt.Sprintf("%s/v1/server/%s", e, api)
}

func (sk SecretKey) hmacSha256(data []byte) []byte {
	hash := hmac.New(sha256.New, sk)
	hash.Write(data)
	return hash.Sum(nil)
}
