# wee

Wee is a powerful and flexible web crawling library based on go-rod, providing a clean interface for writing web crawlers in Go. It offers advanced browser control and customization options.

## Features

- Easy-to-use API for web crawling tasks
- Support for headless and headed browser modes
- User mode for connecting to system Chrome browser
- Cookie management and persistence
- Customizable timeouts and error handling
- Humanized and stealth mode options
- Popover handling
- Advanced browser configuration options
- Support for Chrome extensions
- Proxy configuration

## Installation

```
go get github.com/coghost/wee
```

## Usage

Here's a basic example of using wee to crawl search results from Baidu:

```go
package main

import (
	"fmt"
	"github.com/coghost/wee"
)

func main() {
	bot := wee.NewBotDefault()
	defer bot.BlockInCleanUp()

	const (
		url     = `https://www.baidu.com`
		input   = `input[id="kw"]`
		keyword = `golang`
		items   = `div.result[id] h3>a`
		noItems = `div.this_should_not_existed`
	)

	bot.MustOpen(url)
	bot.MustInput(input, keyword, wee.WithSubmit(true))

	got := bot.MustAnyElem([]string{items, noItems})
	if got == noItems {
		fmt.Printf("%s has no results\n", keyword)
		return
	}

	elems, err := bot.Elems(items)
	if err != nil {
		fmt.Printf("get items failed: %v\n", err)
		return
	}

	for _, elem := range elems {
		fmt.Printf("%s - %s\n", elem.MustText(), *elem.MustAttribute("href"))
	}
}
```

## Advanced Usage

### Creating a Bot

Wee provides several ways to create a bot:

```go
// Default bot
bot := wee.NewBotDefault()

// Headless bot
bot := wee.NewBotHeadless()

// Bot for debugging
bot := wee.NewBotForDebug()

// User mode bot (connects to system Chrome)
bot := wee.NewBotUserMode()
```

### Customizing Bot Behavior

You can customize bot behavior using options:

```go
bot := wee.NewBot(
    wee.UserAgent("Custom User Agent"),
    wee.AcceptLanguage("en-US"),
    wee.WithCookies(true),
    wee.Humanized(true),
    wee.StealthMode(true),
    wee.WithPopovers("popover-selector"),
)
```

### Cookie Management

Wee supports various ways to manage cookies:

```go
// Use cookies from default location
wee.WithCookies(true)

// Specify cookie folder
wee.WithCookieFolder("/path/to/cookies")

// Use specific cookie file
wee.WithCookieFile("/path/to/cookiefile.json")

// Use cookies from clipboard (Copy as cURL)
wee.CopyAsCURLCookies(clipboardContent)
```

### Error Handling

You can customize error handling behavior:

```go
bot.SetPanicWith(wee.PanicByLogError)
```

## Advanced Browser Configuration

Wee provides powerful options to configure the browser:

### Creating a Browser

```go
launcher, browser := wee.NewBrowser(
    wee.BrowserHeadless(false),
    wee.BrowserSlowMotion(1000),
    wee.BrowserUserDataDir("/path/to/user/data"),
    wee.BrowserPaintRects(true),
    wee.BrowserProxy("http://proxy.example.com:8080"),
)
```

### User Mode Browser

```go
launcher, browser := wee.NewUserMode(
    wee.BrowserUserDataDir("/path/to/user/data"),
    wee.LaunchLeakless(true),
)
```

### Customizing Launcher

```go
launcher := wee.NewLauncher(
    wee.BrowserHeadless(true),
    wee.BrowserExtensions("/path/to/extension1", "/path/to/extension2"),
    wee.BrowserFlags("--disable-gpu", "--no-sandbox"),
)
```

## Browser Options

Wee supports various browser options:

- `BrowserHeadless(bool)`: Run browser in headless mode
- `BrowserSlowMotion(int)`: Set slow motion delay in milliseconds
- `BrowserUserDataDir(string)`: Set user data directory
- `BrowserPaintRects(bool)`: Show paint rectangles (useful for debugging)
- `BrowserProxy(string)`: Set proxy server
- `BrowserExtensions(...string)`: Load Chrome extensions
- `BrowserFlags(...string)`: Set additional Chrome flags
- `LaunchLeakless(bool)`: Use leakless mode when launching browser
- `BrowserIncognito(bool)`: Launch browser in incognito mode
- `BrowserIgnoreCertErrors(bool)`: Ignore certificate errors

## More Examples

Check out the [examples folder](https://github.com/coghost/wee/tree/main/examples) for more detailed examples.

## Common Issues

### panic: context canceled

The "context canceled" issue usually occurs when an element is bound with a timeout on `bot.Elem(xxx)`. To resolve this, you can use:

```go
elem.CancelTimeout().Timeout(xxx)
```

This will rewrite the previous timeout value.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[MIT License](LICENSE)
