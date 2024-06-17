package wee

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/gookit/goutil/fsutil"
)

// NewChromeExtension will create two files from line, and save to savePath
//   - background.js
//   - manifest.json
//
// line format is: "host:port:username:password:<OTHER>"
func NewChromeExtension(line, savePath string) (string, error) {
	proxyJS := `var config = {
  mode: 'fixed_servers',
  rules: {
    singleProxy: {
      scheme: 'http',
      host: '%s',
      port: parseInt(%s),
    },
    bypassList: ['foobar.com'],
  },
}
chrome.proxy.settings.set({ value: config, scope: 'regular' }, function () {})
function callbackFn(details) {
  return {
    authCredentials: {
      username: '%s',
      password: '%s',
    },
  }
}
chrome.webRequest.onAuthRequired.addListener(callbackFn, { urls: ['<all_urls>'] }, ['blocking'])`
	manifest := `{
    "version": "1.0.0",
    "manifest_version": 2,
    "name": "GoccProxy",
    "permissions": ["proxy", "tabs", "unlimitedStorage", "storage", "<all_urls>", "webRequest", "webRequestBlocking"],
    "background": {
        "scripts": ["background.js"]
    },
    "minimum_chrome_version": "22.0.0"
}`
	arr := strings.Split(line, ":")
	proxyJS = fmt.Sprintf(proxyJS, arr[0], arr[1], arr[2], arr[3])

	proxyID := arr[0]

	if strings.Contains(arr[0], "superproxy") {
		ipArr := strings.Split(arr[2], "-")
		proxyID = ipArr[len(ipArr)-1]
	}

	proxyID = strings.ReplaceAll(proxyID, ".", "_")

	baseDir := filepath.Join(savePath, proxyID)
	bgJS := filepath.Join(baseDir, "background.js")
	mfJSON := filepath.Join(baseDir, "manifest.json")

	if err := writeFile(bgJS, []byte(proxyJS)); err != nil {
		return "", err
	}

	if err := writeFile(mfJSON, []byte(manifest)); err != nil {
		return "", err
	}

	return baseDir, nil
}

func writeFile(filepath string, content []byte) error {
	var mode fs.FileMode = 0o644

	if err := fsutil.MkParentDir(filepath); err != nil {
		return err
	}

	return os.WriteFile(filepath, content, mode)
}
