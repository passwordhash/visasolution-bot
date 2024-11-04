package config

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

type ProxiesManager struct {
	proxies   []Proxy
	currIndex int
}

type proxiesConfig struct {
	RussianProxies []string `json:"russian_proxies"`
}

func (p *ProxiesManager) Proxies() []Proxy {
	return p.proxies
}

func (p *ProxiesManager) Next() Proxy {
	p.currIndex = (p.currIndex + 1) % len(p.proxies)
	return p.proxies[p.currIndex]
}

func (p *ProxiesManager) Current() Proxy {
	return p.proxies[p.currIndex]
}

type Proxy struct {
	Host     string
	Port     string
	Username string
	Password string
}

// URL возвращает строку вида "http://username:password@host:port"
func (p *Proxy) URL() string {
	return fmt.Sprintf("http://%s:%s@%s:%s", p.Username, p.Password, p.Host, p.Port)
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
func ParseProxiesFile(proxiesFile []byte) (*ProxiesManager, error) {
	var proxies []Proxy
	var proxisConfig proxiesConfig

	err := json.Unmarshal(proxiesFile, &proxisConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse proxies file: %w", err)
	}

	for _, proxyRow := range proxisConfig.RussianProxies {
		proxy, err := ParseProxy(proxyRow)
		if err != nil {
			log.Printf("failed to parse proxy: %v", err)
			continue
		}
		proxies = append(proxies, proxy)
	}

	return &ProxiesManager{proxies: proxies}, nil
}

// ParseProxy приминает прокси в виде "ip:host@usrname:pswrd" и возвращает структуру Proxy
func ParseProxy(proxyRow string) (Proxy, error) {
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
