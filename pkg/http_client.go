package xdeb

import "net/http"

func NewHttpClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			ResponseHeaderTimeout: HTTP_REQUEST_HEADERS_TIMEOUT,
		},
	}
}
