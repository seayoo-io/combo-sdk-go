# Combo SDK for Go

`combo-sdk-go` 是世游核心系统 (Combo) 为 Go 提供的 SDK。

## 服务端

面向服务端提供以下功能，供游戏侧使用：

- 请求 Server REST API 并解析响应
- 接收 Server Notifications 并回复响应
- 验证世游服务端签发的 Identity Token

`combo-sdk-go` 会将 API 的请求响应结构、签名计算与签名验证、HTTP 状态码等实现细节封装起来，提供 Go 的强类型 API，降低游戏侧接入世游系统时出错的可能性，提高接入的速度。
