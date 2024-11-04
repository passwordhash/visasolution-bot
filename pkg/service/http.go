package service

import (
	"net/http"
	"net/url"
)

// ProxyTransport возвращает http.RoundTripper с настроенным прокси
// proxyURL - адрес прокси сервера в формате http://your-proxy-server:port@user:password
func ProxyTransport(proxyURL string) (http.RoundTripper, error) {
	proxy, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}

	return &http.Transport{
		Proxy: http.ProxyURL(proxy),
	}, nil
}
