package combo

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	identityTokenScope = "auth"

	// HMAC-SHA256
	signingMethod = "HS256"
)

// TokenVerifier 用于验证世游服务端颁发的 Token。
type TokenVerifier interface {
	// VerifyIdentityToken 对 IdentityToken 进行验证，返回 IdentityPayload。
	// 如果验证不通过，返回 error。
	VerifyIdentityToken(idToken string) (*IdentityPayload, error)
}

// NewTokenVerifier 创建一个新的 TokenVerifier。
func NewTokenVerifier(o Options) (TokenVerifier, error) {
	if err := o.init(); err != nil {
		return nil, err
	}
	return &verifier{
		parser: jwt.NewParser(
			jwt.WithValidMethods([]string{signingMethod}),
			jwt.WithExpirationRequired(),
			jwt.WithIssuer(string(o.Endpoint)),
			jwt.WithAudience(string(o.GameId)),
		),
		key: o.SecretKey,
	}, nil
}

// IdentityPayload 包含了用户的身份信息。
type IdentityPayload struct {
	// ExpiresAt 是 IdentityToken 的过期时间。
	ExpiresAt time.Time

	// ComboId 是世游分配的聚合用户 ID。
	// 游戏侧应当使用 ComboId 作为用户的唯一标识。
	ComboId string

	// IdP (Identity Provider) 是用户身份的提供者。
	// 游戏侧可以使用 IdP 做业务辅助判断。
	IdP string

	// ExternalId 是用户在外部 IdP 中的唯一标识。
	//
	// 例如：
	//
	// - 如果用户使用世游通行证登录，那么 ExternalId 就是用户的世游通行证 ID。
	//
	// - 如果用户使用 Google Account 登录，那么 ExternalId 就是用户在 Google 中的账号标识。
	//
	// - 如果用户使用微信登录，那么 ExternalId 就是用户在微信中的 OpenId。
	//
	// 注意：游戏侧不应当使用 ExternalId 作为用户标识，但可以使用 ExternalId 做业务辅助判断。
	ExternalId string

	// ExternalName 是用户在外部 IdP 中的名称，通常是用户的昵称。
	ExternalName string

	// WeixinUnionid 是用户在微信中的 UnionId。
	// 游戏侧可以使用 WeixinUnionid 实现多端互通。
	//
	// 注意：WeixinUnionid 只在 IdP 为 weixin 时才会有值。
	WeixinUnionid string
}

type identityClaims struct {
	jwt.RegisteredClaims

	Scope         string `json:"scope"`
	IdP           string `json:"idp"`
	ExternalId    string `json:"external_id"`
	ExternalName  string `json:"external_name"`
	WeixinUnionid string `json:"weixin_unionid"`
}

type verifier struct {
	parser *jwt.Parser
	key    []byte // HS256 signing key
}

func (v *verifier) VerifyIdentityToken(idToken string) (*IdentityPayload, error) {
	token, err := v.parser.ParseWithClaims(idToken, &identityClaims{}, v.keyFunc)
	if err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	claims := token.Claims.(*identityClaims)
	if claims.Scope != identityTokenScope {
		return nil, fmt.Errorf("invalid scope: %s", claims.Scope)
	}
	return &IdentityPayload{
		// claims.ExpiresAt should never be nil
		// because of jwt.WithExpirationRequired() when creating the parser
		ExpiresAt:     claims.ExpiresAt.Time,
		ComboId:       claims.Subject,
		IdP:           claims.IdP,
		ExternalId:    claims.ExternalId,
		ExternalName:  claims.ExternalName,
		WeixinUnionid: claims.WeixinUnionid,
	}, nil
}

func (v *verifier) keyFunc(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return v.key, nil
}

func init() {
	jwt.MarshalSingleStringAsArray = false
}
