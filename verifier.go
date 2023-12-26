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
	// 游戏侧应当使用 ComboId 作为用户的唯一标识。
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
}

// AdPayload 包含了激励广告的播放信息。
type AdPayload struct {
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
