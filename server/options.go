package server

import (
	"net/http"

	"github.com/seayoo-io/combo-sdk-go/combo"
)

// HttpClient provides the interface to provide custom HTTPClients.
// Generally *http.Client is sufficient for most use cases.
type HttpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Options struct {
	// (Required) The endpoint to send requests to.
	Endpoint combo.Endpoint

	// (Required) The Game ID to use for requests.
	GameId combo.GameId

	// (Required) The Secret Key to use for requests.
	SecretKey combo.SecretKey

	// (Optional) The HTTP client to use when sending requests.
	// If nil, *http.Client will be created and used.
	HttpClient HttpClient

	// (Optional) The HttpSigner to use when signing requests.
	// If nil, a default HttpSigner will be created and used.
	Signer HttpSigner
}
