package worker

import (
	"fmt"
	"strings"
	"visasolution/pkg/util"
)

const chromeExtensionFilename = "proxy_auth_plugin.zip"

var chromeExtensionRelativePath = tmpFolder + chromeExtensionFilename

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

// FromRow приминает прокси в виде "ip:host@usrname:pswrd"
func FromRow(proxyRow string) (Proxy, error) {
	var proxy Proxy

	parts := strings.Split(proxyRow, "@")
	if len(parts) != 2 {
		return proxy, fmt.Errorf("invalid proxy format, expected 'ip:port@username:password'")
	}

	ipPort := parts[0]
	auth := parts[1]

	ipPortParts := strings.Split(ipPort, ":")
	if len(ipPortParts) != 2 {
		return proxy, fmt.Errorf("invalid ip:port format")
	}

	proxy.Host = ipPortParts[0]
	proxy.Port = ipPortParts[1]

	authParts := strings.Split(auth, ":")
	if len(authParts) != 2 {
		return proxy, fmt.Errorf("invalid username:password format")
	}

	proxy.Username = authParts[0]
	proxy.Password = authParts[1]

	return proxy, nil
}

// GenerateProxyAuthExtension приминает прокси в виде "ip:host@usrname:pswrd"
func (w *Worker) GenerateProxyAuthExtension(proxyRow string) (string, error) {
	proxy, err := FromRow(proxyRow)
	if err != nil {
		return "", err
	}

	filenames := []string{"manifest.json", "background.js"}
	manifestContent := []byte(fmt.Sprintf(manifest))
	backgroundJSContent := []byte(fmt.Sprintf(backgroundJS, proxy.Host, proxy.Port, proxy.Username, proxy.Password))

	err = util.CreateZip(filenames, [][]byte{manifestContent, backgroundJSContent}, chromeExtensionRelativePath)
	if err != nil {
		return "", fmt.Errorf("error creating ZIP file: %v", err)
	}

	return chromeExtensionRelativePath, nil
}
