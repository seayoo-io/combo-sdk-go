package combo

import (
	"net/http"
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
	// If nil, *http.Client will be created and used.
	HttpClient HttpClient

	// (Optional) The HttpSigner to use when signing requests.
	// If nil, a default HttpSigner will be created and used.
	HttpSigner HttpSigner
}
