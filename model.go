package combo

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

var (
	// 中国大陆 API 端点，用于国内发行
	EndpointChina Endpoint = Endpoint{Url: "https://api.seayoo.com"}

	// 全球的 API 端点，用于海外发行
	EndpointGlobal Endpoint = Endpoint{Url: "https://api.seayoo.io"}
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

type Platform string

type Endpoint struct {
	Url string
}

type GameId struct {
	Id string
}

type SecretKey struct {
	Key []byte
}

func (e Endpoint) String() string {
	return e.Url
}

func (e Endpoint) apiUrl(api string) string {
	return fmt.Sprintf("%s/v3/server/%s", e.Url, api)
}

func (gid GameId) String() string {
	return gid.Id
}

func (sk SecretKey) String() string {
	return string(sk.Key)
}

func (sk SecretKey) HmacSha256(data []byte) []byte {
	hash := hmac.New(sha256.New, sk.Key)
	hash.Write(data)
	return hash.Sum(nil)
}
