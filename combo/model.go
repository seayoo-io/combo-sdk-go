package combo

import (
	"crypto/hmac"
	"crypto/sha256"
)

type GameID struct {
	ID string
}

type SecretKey struct {
	Key []byte
}

func (gid GameID) String() string {
	return gid.ID
}

func (sk SecretKey) String() string {
	return string(sk.Key)
}

func (sk SecretKey) HMACSHA256(data []byte) []byte {
	hash := hmac.New(sha256.New, sk.Key)
	hash.Write(data)
	return hash.Sum(nil)
}
