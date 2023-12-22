package combo

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

const (
	// 中国大陆 API 端点，用于国内发行
	EndpointChina Endpoint = "https://api.seayoo.com"

	// 全球的 API 端点，用于海外发行
	EndpointGlobal Endpoint = "https://api.seayoo.io"
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
)

// 游戏客户端运行平台
type Platform string

// Combo API 端点
type Endpoint string

// 由世游为游戏分配，用于标识游戏的业务代号。
type GameId string

// 由世游侧为游戏分配，游戏侧和世游侧共享的密钥。
// 此密钥用于签名计算与验证。
type SecretKey []byte

func (e Endpoint) String() string {
	return string(e)
}

func (e Endpoint) url(api string) string {
	return fmt.Sprintf("%s/v3/server/%s", e, api)
}

func (gid GameId) String() string {
	return string(gid)
}

func (sk SecretKey) String() string {
	return string(sk)
}

func (sk SecretKey) hmacSha256(data []byte) []byte {
	hash := hmac.New(sha256.New, sk)
	hash.Write(data)
	return hash.Sum(nil)
}
