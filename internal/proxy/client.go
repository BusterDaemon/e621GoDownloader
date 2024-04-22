package proxy

import (
	"net/http"
	"net/url"
	"time"
)

func NewClient(proxy string) *http.Client {
	proxyUrl, err := url.Parse(proxy)
	if err != nil {
		panic(err)
	}

	client := http.DefaultClient
	client.Transport = DefaultTransport(proxyUrl)

	return client
}

func DefaultTransport(proxyUrl *url.URL) *http.Transport {
	return &http.Transport{
		Proxy:                 http.ProxyURL(proxyUrl),
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
