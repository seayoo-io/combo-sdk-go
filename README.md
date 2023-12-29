# Combo SDK for Go

`combo-sdk-go` 是世游核心系统 (Combo) 为 Go 提供的 SDK。

提供以下服务端功能，供游戏侧使用：

- 验证世游服务端签发的 Identity Token
- 请求 Server REST API 并解析响应
- 接收 Server Notifications 并回复响应

`combo-sdk-go` 会将 API 的请求响应结构、签名计算与签名验证、HTTP 状态码等实现细节封装起来，提供 Go 的强类型 API，降低游戏侧接入世游系统时出错的可能性，提高接入的速度。

## API Reference

- https://pkg.go.dev/github.com/seayoo-io/combo-sdk-go

## 初始化

```go
package main

import "github.com/seayoo-io/combo-sdk-go"

func main() {
	cfg := combo.Config{
		Endpoint:  combo.Endpoint_China, // or combo.Endpoint_Global
		GameId:    combo.GameId("<GAME_ID>"),
		SecretKey: combo.SecretKey("sk_<SECRET_KEY>"),
	}
	// Use cfg...
}
```

## 登录验证

```go
package main

import (
	"fmt"

	"github.com/seayoo-io/combo-sdk-go"
)

func main() {
	cfg := combo.Config{
		Endpoint:  combo.Endpoint_China, // or combo.Endpoint_Global
		GameId:    combo.GameId("<GAME_ID>"),
		SecretKey: combo.SecretKey("sk_<SECRET_KEY>"),
	}

	verifier, err := combo.NewTokenVerifier(cfg)
	if err != nil {
		panic(err)
	}

	// TokenVerifier 是可以复用的，不需要每次验证 Token 都创建一个 TokenVerifier。
	VerifyIdentityToken(verifier, "<IDENTITY_TOKEN_1>")
	VerifyIdentityToken(verifier, "<IDENTITY_TOKEN_2>")
	VerifyIdentityToken(verifier, "<IDENTITY_TOKEN_3>")
}

func VerifyIdentityToken(verifier *combo.TokenVerifier, token string) {
	payload, err := verifier.VerifyIdentityToken(token)
	if err != nil {
		fmt.Printf("failed to verify identity token: %v\n", err)
		return
	}
	fmt.Printf("ComboId: %s\n", payload.ComboId)
	fmt.Printf("IdP: %s\n", payload.IdP)
	fmt.Printf("ExternalId: %s\n", payload.ExternalId)
	fmt.Printf("ExternalName: %s\n", payload.ExternalName)
}
```

## 创建订单

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/seayoo-io/combo-sdk-go"
)

func main() {
	cfg := combo.Config{
		Endpoint:  combo.Endpoint_China, // or combo.Endpoint_Global
		GameId:    combo.GameId("<GAME_ID>"),
		SecretKey: combo.SecretKey("sk_<SECRET_KEY>"),
	}

	client, err := combo.NewClient(cfg)
	if err != nil {
		panic(err)
	}

	// Client 是可以复用的，不需要每次请求都创建一个新的 Client
	output, err := client.CreateOrder(context.Background(), &combo.CreateOrderInput{
		Platform:    combo.Platform_iOS,
		ReferenceId: "20b5f268-22e1-4677-9a3f-ed8c9f22ff0f",
		ComboId:     "1231229080370001",
		ProductId:   "xcom_product_648",
		Quantity:    1,
		NotifyUrl:   "https://example.com/notifications",
		Context:     "<WHAT_YOU_PUT_HERE_IS_WHAT_YOU_GET_IN_NOTIFICATION>",
		Meta: combo.OrderMeta{
			ZoneId:    "10000",
			ServerId:  "10001",
			RoleId:    "3888",
			RoleName:  "小明",
			RoleLevel: 59,
		},
	})

	if err != nil {
		var er *combo.ErrorResponse
		if errors.As(err, &er) {
			fmt.Println("failed to create order, got ErrorResponse:")
			fmt.Printf("StatusCode: %d\n", er.StatusCode())
			fmt.Printf("TraceId: %s\n", er.TraceId())
			fmt.Printf("ErrorCode: %s\n", er.ErrorCode)
			fmt.Printf("ErrorMessage: %s\n", er.ErrorMessage)
		} else {
			fmt.Printf("failed to create order: %v\n", err)
		}
		return
	}

	fmt.Println("successfully created order:")
	fmt.Printf("TraceId: %s\n", output.TraceId())
	fmt.Printf("OrderId: %s\n", output.OrderId)
	fmt.Printf("OrderToken: %s\n", output.OrderToken)
	fmt.Printf("ExpiresAt: %v\n", time.Unix(output.ExpiresAt, 0))
}
```

## 处理发货通知

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/seayoo-io/combo-sdk-go"
)

func main() {
	cfg := combo.Config{
		Endpoint:  combo.Endpoint_China, // or combo.Endpoint_Global
		GameId:    combo.GameId("<GAME_ID>"),
		SecretKey: combo.SecretKey("sk_<SECRET_KEY>"),
	}

	handler, err := combo.NewNotificationHandler(cfg, &NotificationListener{})
	if err != nil {
		panic(err)
	}
	http.Handle("/notifications", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type NotificationListener struct{}

func (l *NotificationListener) HandleShipOrder(ctx context.Context, id combo.NotificationId, payload *combo.ShipOrderNotification) error {
	fmt.Printf("received ship order notification: %s\n", id)
	fmt.Printf("OrderId: %s\n", payload.OrderId)
	fmt.Printf("ReferenceId: %s\n", payload.ReferenceId)
	fmt.Printf("ComboId: %s\n", payload.ComboId)
	fmt.Printf("ProductId: %s\n", payload.ProductId)
	fmt.Printf("Quantity: %d\n", payload.Quantity)
	fmt.Printf("Currency: %s\n", payload.Currency)
	fmt.Printf("ComboId: %d\n", payload.Amount)
	fmt.Printf("Context: %s\n", payload.Context)
	return nil
}
```
