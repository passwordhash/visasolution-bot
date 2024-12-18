package worker

import (
	"fmt"
	"path"
	"visasolution/internal/config"
	"visasolution/pkg/util"
)

const chromeExtensionFilename = "proxy_auth_plugin.zip"

// manifest шаблон для файла манифеста расширения
const manifest = `
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
const backgroundJS = `
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

// GenerateProxyAuthExtension приминает прокси в виде "ip:host@usrname:pswrd"
func (w *Worker) GenerateProxyAuthExtension(proxy config.Proxy) (string, error) {
	extensionPath := w.chromeExtensionPath()

	filenames := []string{"manifest.json", "background.js"}
	manifestContent := []byte(fmt.Sprintf(manifest))
	backgroundJSContent := []byte(fmt.Sprintf(backgroundJS, proxy.Host, proxy.Port, proxy.Username, proxy.Password))

	err := util.CreateZip(filenames, [][]byte{manifestContent, backgroundJSContent}, extensionPath)
	if err != nil {
		return "", fmt.Errorf("error creating ZIP file: %v", err)
	}

	return extensionPath, nil
}

func (w *Worker) chromeExtensionPath() string {
	return path.Join(w.d.TmpFolder, chromeExtensionFilename)
}
