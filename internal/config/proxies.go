package config

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

type Proxy struct {
	Host     string
	Port     string
	Username string
	Password string
}

type proxiesConfig struct {
	RussianProxies []string `json:"russian_proxies"`
}

// ParseProxiesFile принимает содержимое файла с прокси и возвращает слайс из Proxy.
// Параметры:
// - proxiesFile содержимое файла с прокси в формате JSON. Пример содержимого:
//
//	{
//	  "russian_proxies": [
//	    "ip:host@usrname:pswrd"
//	  ]
//	}
func ParseProxiesFile(proxiesFile []byte) ([]Proxy, error) {
	var proxies []Proxy
	var proxisConfig proxiesConfig

	err := json.Unmarshal(proxiesFile, &proxisConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse proxies file: %w", err)
	}

	for _, proxyRow := range proxisConfig.RussianProxies {
		proxy, err := parseProxy(proxyRow)
		if err != nil {
			log.Printf("failed to parse proxy: %v", err)
			continue
		}
		proxies = append(proxies, proxy)
	}

	return proxies, nil
}

// parseProxy приминает прокси в виде "ip:host@usrname:pswrd" и возвращает структуру Proxy
func parseProxy(proxyRow string) (Proxy, error) {
	parts := strings.Split(proxyRow, "@")
	if len(parts) != 2 {
		return Proxy{}, fmt.Errorf("invalid proxy format, expected 'ip:port@username:password'")
	}

	ipPortParts := strings.Split(parts[0], ":")
	authParts := strings.Split(parts[1], ":")
	if len(ipPortParts) != 2 || len(authParts) != 2 {
		return Proxy{}, fmt.Errorf("invalid format")
	}

	return Proxy{
		Host:     ipPortParts[0],
		Port:     ipPortParts[1],
		Username: authParts[0],
		Password: authParts[1],
	}, nil
}
