package combo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	notificationType_ShipOrder = "ship_order"
)

// NewNotificationHandler 创建一个用于接收世游服务端推送的通知的 http.Handler。
//
// 游戏侧需要将此 Handler 注册到游戏的 HTTP 服务中。
//
// 注意：注册 Handler 时，应当使用 HTTP POST。
func NewNotificationHandler(cfg Config, listener NotificationListener) (http.Handler, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	if listener == nil {
		return nil, errors.New("missing required listener")
	}
	return &notificationHandler{
		signer: httpSigner{
			game:       cfg.GameId,
			signingKey: cfg.SecretKey,
		},
		listener: listener,
	}, nil
}

// 每次通知的唯一 ID。游戏侧可用此值来对通知进行去重。
type NotificationId string

// NotificationListener 是一个用于接收世游服务端推送的通知的接口。
//
// 游戏侧需要实现此接口并执行对应的业务逻辑。
type NotificationListener interface {
	HandleShipOrder(ctx context.Context, id NotificationId, payload *ShipOrderNotification) error
}

type ShipOrderNotification struct {
	// 世游服务端创建的,标识订单的唯一 ID。
	OrderId string `json:"order_id"`

	// 游戏侧用于标识创建订单请求的唯一 ID。
	ReferenceId string `json:"reference_id"`

	// 发起购买的用户的唯一标识。
	ComboId string `json:"combo_id"`

	// 购买的商品 ID。
	ProductId string `json:"product_id"`

	// 购买的商品的数量。
	Quantity int `json:"quantity"`

	// 订单币种代码。例如 USD CNY。
	Currency string `json:"currency"`

	// 订单金额,单位为分。
	Amount int `json:"amount"`

	// 游戏侧创建订单时提供的订单上下文，透传回游戏。
	Context string `json:"context"`
}

type notificationRequestBody struct {
	// 世游服务端通知的版本号。当前版本固定为 1.0。
	Version string `json:"version"`

	// 每次通知的唯一 ID。游戏侧可用此值来对通知进行去重。
	Id string `json:"notification_id"`

	// 用于标识通知类型，Data 中的结构随着通知类型的不同而不同。
	Type string `json:"notification_type"`

	// 与 Type 对应的结构，是一个异构的 JSON Object。
	Data json.RawMessage `json:"data"`
}

type notificationHandler struct {
	signer   httpSigner
	listener NotificationListener
}

func (h *notificationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "please use HTTP POST", http.StatusMethodNotAllowed)
		return
	}
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		http.Error(w, "please use application/json", http.StatusUnsupportedMediaType)
		return
	}
	if err := h.signer.AuthHttp(r, time.Now()); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	var body notificationRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	switch body.Type {
	case notificationType_ShipOrder:
		var payload ShipOrderNotification
		if err := json.Unmarshal(body.Data, &payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := h.listener.HandleShipOrder(r.Context(), NotificationId(body.Id), &payload); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, fmt.Sprintf("unknown notification type: %s", body.Type), http.StatusBadRequest)
		return
	}
	// Notification has been handled successfully, return 200 OK
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
