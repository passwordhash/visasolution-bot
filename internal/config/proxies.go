package config

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

type ProxiesManager struct {
	proxiesRU []Proxy
	currIndex int

	// ProxyForeign - прокси для иностранных сайтов.
	// Может быть nil, если не указан в конфиге.
	ProxyForeign Proxy
}

type proxiesConfig struct {
	RussianProxies []string `json:"russian_proxies"`
	ForeignProxy   string   `json:"foreign_proxy"`
}

func (p *ProxiesManager) ProxiesRU() []Proxy {
	return p.proxiesRU
}

func (p *ProxiesManager) NextRU() Proxy {
	p.currIndex = (p.currIndex + 1) % len(p.proxiesRU)
	return p.proxiesRU[p.currIndex]
}

func (p *ProxiesManager) CurrentRU() Proxy {
	return p.proxiesRU[p.currIndex]
}

// Proxy - структура для хранения авторизационных данных прокси
type Proxy struct {
	Host     string
	Port     string
	Username string
	Password string
}

func (p *Proxy) IsEmpty() bool {
	return p.Host == "" || p.Port == "" || p.Username == "" || p.Password == ""
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
//	  ],
//	  "foreign_proxy": "ip:host@usrname:pswrd"
//	}
func ParseProxiesFile(proxiesFile []byte) (*ProxiesManager, error) {
	var proxisConfig proxiesConfig
	var proxiesRU []Proxy

	err := json.Unmarshal(proxiesFile, &proxisConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse proxiesRU file: %w", err)
	}

	for _, proxyRow := range proxisConfig.RussianProxies {
		proxy, err := ParseProxy(proxyRow)
		if err != nil {
			log.Printf("failed to parse proxy: %v", err)
			continue
		}
		proxiesRU = append(proxiesRU, proxy)
	}

	proxyForeign, _ := ParseProxy(proxisConfig.ForeignProxy)

	return &ProxiesManager{proxiesRU: proxiesRU, ProxyForeign: proxyForeign}, nil
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
