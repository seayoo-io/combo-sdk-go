package combo

import (
	"crypto/hmac"
	"crypto/sha256"
)

type GameId struct {
	Id string
}

type SecretKey struct {
	Key []byte
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
