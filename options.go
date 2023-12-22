package combo

import (
	"errors"
	"net/http"
	"strings"
)

// HttpClient provides the interface to provide custom HTTPClients.
// Generally *http.Client is sufficient for most use cases.
type HttpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Options struct {
	// (Required) The endpoint to send requests to.
	Endpoint Endpoint

	// (Required) The Game ID to use for requests.
	GameId GameId

	// (Required) The Secret Key to use for requests.
	SecretKey SecretKey

	// (Optional) The HTTP client to use when sending requests.
	// If nil, http.DefaultClient will be used.
	HttpClient HttpClient

	// (Optional) The HttpSigner to use when signing requests.
	// If nil, a default HttpSigner will be created and used.
	HttpSigner HttpSigner
}

func (o *Options) init() error {
	if o.Endpoint == "" {
		return errors.New("missing required Endpoint")
	}
	o.Endpoint = Endpoint(strings.TrimSuffix(string(o.Endpoint), "/"))
	if o.GameId == "" {
		return errors.New("missing required GameId")
	}
	if o.SecretKey == nil || len(o.SecretKey) == 0 {
		return errors.New("missing required SecretKey")
	}
	if !strings.HasPrefix(string(o.SecretKey), "sk_") {
		return errors.New("invalid SecretKey: must start with sk_")
	}
	if o.HttpClient == nil {
		o.HttpClient = http.DefaultClient
	}
	if o.HttpSigner == nil {
		signer, err := NewHttpSigner(o.GameId, o.SecretKey)
		if err != nil {
			return err
		}
		o.HttpSigner = signer
	}
	return nil
}
