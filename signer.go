package combo

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	signingAlgorithm    = "SEAYOO-HMAC-SHA256"
	authorizationHeader = "Authorization"

	// timeFormat is the ISO 8601 format string for the signing timestamp
	timeFormat = "20060102T150405Z"

	// emptyStringSha256 is the hex encoded sha256 value of an empty string
	emptyStringSha256 = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	// maxTimeDiff is the maximum allowed time difference between the signing time and the current time
	maxTimeDiff = time.Minute * 5
)

// HttpSigner is the interface for signing and verifying http requests
type HttpSigner interface {
	// SignHttp computes the signature of given http request and set the Authorization header
	SignHttp(r *http.Request, signingTime time.Time) error

	// AuthHttp reads the Authorization header from given http request and verifies the signature
	AuthHttp(r *http.Request, currentTime time.Time) error
}

// NewHttpSigner creates a new HTTPSigner
func NewHttpSigner(gameId GameId, secretKey SecretKey) (HttpSigner, error) {
	return &signer{
		gameId:    gameId,
		secretKey: secretKey,
	}, nil
}

type signer struct {
	gameId    GameId
	secretKey SecretKey
}

type authorization struct {
	scheme    string
	game      GameId
	timestamp time.Time
	signature string
}

func (s *signer) SignHttp(r *http.Request, signingTime time.Time) error {
	timestamp := getTimestamp(signingTime)
	stringToSign, err := buildStringToSign(r, timestamp)
	if err != nil {
		return err
	}
	signature := s.computeSignature(stringToSign)
	r.Header.Set(authorizationHeader, s.buildAuthorizationHeader(timestamp, signature))
	return nil
}

func (s *signer) AuthHttp(r *http.Request, currentTime time.Time) error {
	// Step 1, parse authorization header
	auth, err := parseAuthorizationHeader(r.Header.Get(authorizationHeader))
	if err != nil {
		return err
	}
	// Step 2, verify scheme
	if auth.scheme != signingAlgorithm {
		return fmt.Errorf("invalid auth scheme: %s", auth.scheme)
	}
	// Step 3, verify timestamp
	timeDiff := currentTime.Sub(auth.timestamp).Abs()
	if timeDiff > maxTimeDiff {
		return fmt.Errorf("time difference exceeds maximum allowed: %s", timeDiff)
	}
	// Step 4, verify game
	if auth.game != s.gameId {
		return fmt.Errorf("invalid game: %s", auth.game)
	}
	// Step 5, verify signature
	timestamp := getTimestamp(auth.timestamp)
	stringToSign, err := buildStringToSign(r, timestamp)
	if err != nil {
		return err
	}
	signature := s.computeSignature(stringToSign)
	if auth.signature != signature {
		return fmt.Errorf("invalid signature: expect %s, got %s", signature, auth.signature)
	}
	return nil
}

func (s *signer) computeSignature(stringToSign string) string {
	sig := s.secretKey.hmacSha256([]byte(stringToSign))
	return hex.EncodeToString(sig)
}

func (s *signer) buildAuthorizationHeader(timestamp, signature string) string {
	// TODO: include space between parameters
	// return fmt.Sprintf("%s Game=%s, Timestamp=%s, Signature=%s",
	return fmt.Sprintf("%s Game=%s,Timestamp=%s,Signature=%s",
		signingAlgorithm,
		s.gameId,
		timestamp,
		signature,
	)
}

func getTimestamp(t time.Time) string {
	return t.UTC().Format(timeFormat)
}

func buildStringToSign(r *http.Request, timestamp string) (string, error) {
	payloadHash, err := computePayloadHash(r)
	if err != nil {
		return "", err
	}
	return strings.Join([]string{
		signingAlgorithm,
		r.Method,
		r.URL.RequestURI(),
		timestamp,
		payloadHash,
	}, "\n"), nil
}

func computePayloadHash(r *http.Request) (string, error) {
	if r.Body == nil {
		return emptyStringSha256, nil
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", fmt.Errorf("failed to compute payload hash: %w", err)
	}
	// Replace the body with a new reader, because we already read it
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	sum256 := sha256.Sum256(body)
	return hex.EncodeToString(sum256[:]), nil
}

// parseAuthorizationHeader parses the Authorization header of a http request according to:
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Authorization
func parseAuthorizationHeader(header string) (*authorization, error) {
	if header == "" {
		return nil, errors.New("missing authorization header")
	}
	space := strings.IndexByte(header, ' ')
	if space == -1 {
		return nil, errors.New("invalid authorization header")
	}
	scheme := header[:space]
	parameters := strings.Split(header[space:], ",")
	auth := authorization{
		scheme: scheme,
	}
	for _, p := range parameters {
		kv := strings.Split(p, "=")
		if len(kv) != 2 {
			return nil, errors.New("invalid parameters in authorization header")
		}
		kv[0] = strings.Trim(kv[0], ` `)
		kv[1] = strings.Trim(kv[1], ` "`)
		switch kv[0] {
		case "Game":
			auth.game = GameId(kv[1])
		case "Timestamp":
			t, err := time.Parse(timeFormat, kv[1])
			if err != nil {
				return nil, fmt.Errorf("invalid timestamp: %w", err)
			}
			auth.timestamp = t
		case "Signature":
			auth.signature = kv[1]
		default:
			// Ignore unknown authorization parameters
		}
	}
	return &auth, nil
}
