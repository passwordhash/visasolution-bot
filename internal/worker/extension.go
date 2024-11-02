package worker

import (
	"fmt"
	"strings"
	"visasolution/pkg/util"
)

const chromeExtensionFilename = "proxy_auth_plugin.zip"

// manifest шаблон для файла манифеста расширения
var manifest = `
{
    "version": "1.0.0",
    "manifest_version": 2,
    "name": "Chrome Proxy",
    "permissions": [
        "proxy",
        "tabs",
        "unlimitedStorage",
        "storage",
        "<all_urls>",
        "webRequest",
        "webRequestBlocking"
    ],
    "background": {
        "scripts": ["background.js"]
    },
    "minimum_chrome_version":"22.0.0"
}
`

// backgroundJS шаблон для скрипта расширения
var backgroundJS = `
var config = {
        mode: "fixed_servers",
        rules: {
        singleProxy: {
            scheme: "http",
            host: "%s",
            port: parseInt(%s)
        },
        bypassList: ["localhost"]
        }
    };

chrome.proxy.settings.set({value: config, scope: "regular"}, function() {});

function callbackFn(details) {
    return {
        authCredentials: {
            username: "%s",
            password: "%s"
        }
    };
}

chrome.webRequest.onAuthRequired.addListener(
            callbackFn,
            {urls: ["<all_urls>"]},
            ['blocking']
);
`

type Proxy struct {
	Host     string
	Port     string
	Username string
	Password string
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

// GenerateProxyAuthExtension приминает прокси в виде "ip:host@usrname:pswrd"
func (w *Worker) GenerateProxyAuthExtension(proxyRow string) (string, error) {
	path := w.chromeExtensionPath()
	proxy, err := ParseProxy(proxyRow)
	if err != nil {
		return "", err
	}

	filenames := []string{"manifest.json", "background.js"}
	manifestContent := []byte(fmt.Sprintf(manifest))
	backgroundJSContent := []byte(fmt.Sprintf(backgroundJS, proxy.Host, proxy.Port, proxy.Username, proxy.Password))

	err = util.CreateZip(filenames, [][]byte{manifestContent, backgroundJSContent}, path)
	if err != nil {
		return "", fmt.Errorf("error creating ZIP file: %v", err)
	}

	return path, nil
}

func (w *Worker) chromeExtensionPath() string {
	return w.d.TmpFolder + chromeExtensionFilename
}
