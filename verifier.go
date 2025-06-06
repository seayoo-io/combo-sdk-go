package combo

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

const (
	identityTokenScope = "auth"
	adTokenScope       = "ads"
)

// TokenVerifier 用于验证世游服务端颁发的 Token。
type TokenVerifier struct {
	parser *jwt.Parser
	key    SecretKey
}

// NewTokenVerifier 创建一个新的 TokenVerifier。
func NewTokenVerifier(cfg Config) (*TokenVerifier, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &TokenVerifier{
		parser: jwt.NewParser(
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
			jwt.WithExpirationRequired(),
			jwt.WithIssuer(string(cfg.Endpoint)),
			jwt.WithAudience(string(cfg.GameId)),
		),
		key: cfg.SecretKey,
	}, nil
}

// IdentityPayload 包含了用户的身份信息。
type IdentityPayload struct {
	// ComboId 是世游分配的聚合用户 ID。
	// 游戏侧应当使用 ComboId 作为用户的唯一标识，即游戏帐号。
	ComboId string

	// IdP (Identity Provider) 是用户身份的提供者。
	// 游戏侧可以使用 IdP 做业务辅助判断，例如判定用户是否使用了某个特定的登录方式。
	IdP IdP

	// ExternalId 是用户在外部 IdP 中的唯一标识。
	//
	// 例如：
	// - 如果用户使用世游通行证登录，那么 ExternalId 就是用户的世游通行证 ID。
	// - 如果用户使用 Google Account 登录，那么 ExternalId 就是用户在 Google 中的账号标识。
	// - 如果用户使用微信登录，那么 ExternalId 就是用户在微信中的 OpenId。
	//
	// 注意：
	// 游戏侧不应当使用 ExternalId 作为用户标识，但可以将 ExternalId 用于特定的业务逻辑。
	ExternalId string

	// ExternalName 是用户在外部 IdP 中的名称，通常是用户的昵称。
	ExternalName string

	// WeixinUnionid 是用户在微信中的 UnionId。
	// 游戏侧可以使用 WeixinUnionid 实现多端互通。
	//
	// 注意：WeixinUnionid 只在 IdP 为 weixin 时才会有值。
	WeixinUnionid string

	// DeviceId 是用户在登录时使用的设备的唯一 ID。
	DeviceId string

	// Distro 是游戏客户端的发行版本标识。
	// 游戏侧可将 Distro 用于服务端数据埋点，以及特定的业务逻辑判断。
	Distro string

	// Variant 是游戏客户端的分包标识。
	// 游戏侧可将 Variant 用于服务端数据埋点，以及特定的业务逻辑判断。
	//
	// 注意：Variant 只在客户端是分包时才会有值。当客户端不是分包的情况下，Variant 为空字符串。
	Variant string

	// Age 是根据用户的实名认证信息得到的年龄。0 表示未知。
	//
	// 在某些特殊场景下，游戏侧可用 Age 来自行处理防沉迷。
	//
	// 注意：Age 不保证返回精确的年龄信息，仅保证用于防沉迷处理时的准确度够用。
	// 例如：
	//	当某个用户真实年龄为 35 岁时，Age 可能返回 18
	//	当某个用户真实年龄为 17 岁时，Age 可能返回 16
	Age int
}

// AdPayload 包含了激励广告的播放信息。
type AdPayload struct {
	// ComboId 是世游分配的聚合用户 ID。
	// 游戏侧应当使用 ComboId 作为用户的唯一标识。
	ComboId string

	// PlacementId 是广告位 ID，游戏侧用它确定发放什么样的广告激励。
	PlacementId string

	// ImpressionId 是世游服务端创建的，标识单次广告播放的唯一 ID。
	ImpressionId string
}

type identityClaims struct {
	jwt.RegisteredClaims
	Scope         string `json:"scope"`
	IdP           string `json:"idp"`
	ExternalId    string `json:"external_id"`
	ExternalName  string `json:"external_name"`
	WeixinUnionid string `json:"weixin_unionid"`
	DeviceId      string `json:"device_id"`
	Distro        string `json:"distro"`
	Variant       string `json:"variant"`
	Age           int    `json:"age"`
}

type adClaims struct {
	jwt.RegisteredClaims
	Scope        string `json:"scope"`
	PlacementId  string `json:"placement_id"`
	ImpressionId string `json:"impression_id"`
}

// VerifyIdentityToken 对 IdentityToken 进行验证。
//
// 如果验证通过，返回 IdentityPayload。如果验证不通过，返回 error。
func (v *TokenVerifier) VerifyIdentityToken(tokenString string) (*IdentityPayload, error) {
	token, err := v.parseToken(tokenString, &identityClaims{})
	if err != nil {
		return nil, err
	}
	claims := token.Claims.(*identityClaims)
	if claims.Scope != identityTokenScope {
		return nil, fmt.Errorf("invalid scope: %s", claims.Scope)
	}
	return &IdentityPayload{
		ComboId:       claims.Subject,
		IdP:           IdP(claims.IdP),
		ExternalId:    claims.ExternalId,
		ExternalName:  claims.ExternalName,
		WeixinUnionid: claims.WeixinUnionid,
		DeviceId:      claims.DeviceId,
		Distro:        claims.Distro,
		Variant:       claims.Variant,
		Age:           claims.Age,
	}, nil
}

// VerifyAdToken 对 AdToken 进行验证。
//
// 如果验证通过，返回 AdPayload。如果验证不通过，返回 error。
func (v *TokenVerifier) VerifyAdToken(tokenString string) (*AdPayload, error) {
	token, err := v.parseToken(tokenString, &adClaims{})
	if err != nil {
		return nil, err
	}
	claims := token.Claims.(*adClaims)
	if claims.Scope != adTokenScope {
		return nil, fmt.Errorf("invalid scope: %s", claims.Scope)
	}
	return &AdPayload{
		ComboId:      claims.Subject,
		PlacementId:  claims.PlacementId,
		ImpressionId: claims.ImpressionId,
	}, nil
}

func (v *TokenVerifier) parseToken(tokenString string, claims jwt.Claims) (*jwt.Token, error) {
	token, err := v.parser.ParseWithClaims(tokenString, claims, v.keyFunc)
	if err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return token, nil
}

func (v *TokenVerifier) keyFunc(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return []byte(v.key), nil
}

func init() {
	jwt.MarshalSingleStringAsArray = false
}
