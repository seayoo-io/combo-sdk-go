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

type TokenVerifier interface {
	VerifyIdentityToken(idToken string) (*IdentityPayload, error)
}

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

type IdentityPayload struct {
	ExpiresAt     time.Time
	ComboId       string
	IdP           string
	ExternalId    string
	ExternalName  string
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
