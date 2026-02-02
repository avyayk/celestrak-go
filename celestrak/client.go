package celestrak

import (
	"net/http"
	"time"
)

type CelestrakClient struct {
	baseURL string
	http *http.Client
}

func NewCelestrakClient() *CelestrakClient {
	return &CelestrakClient{
		baseURL: "https://celestrak.org",
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}