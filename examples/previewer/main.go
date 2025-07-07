package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/rod/lib/utils"
	"github.com/ysmood/gson"
)

const (
	cacheDir             = "/tmp/cache"
	previewTimeout       = 5 * time.Second
	maxPreviewsPerMinute = 30
)

func main() {
	ensureCacheDir()
	wsURL := launcher.NewUserMode().MustLaunch()
	browser := rod.New().ControlURL(wsURL).MustConnect().NoDefaultDevice()
	linkPreviewer(browser)

	// Create a test page
	page := browser.MustPage()
	page.MustNavigate("about:blank")

	waitExit()
}

func linkPreviewer(browser *rod.Browser) {
	previewer := rod.New().MustConnect()
	previewer.MustSetCookies(browser.MustGetCookies()...)
	pool := rod.NewPagePool(5)
	create := func() *rod.Page { return previewer.MustPage() }

	// Rate limiter
	rateLimiter := make(chan struct{}, maxPreviewsPerMinute)
	go func() {
		for {
			time.Sleep(time.Minute / maxPreviewsPerMinute)
			select {
			case rateLimiter <- struct{}{}:
			default:
			}
		}
	}()

	go browser.EachEvent(func(e *proto.TargetTargetCreated) {
		if e.TargetInfo.Type != proto.TargetTargetInfoTypePage {
			return
		}
		page := browser.MustPageFromTargetID(e.TargetInfo.TargetID)
		page.MustEvalOnNewDocument(getJSCode())

		page.MustExpose("getPreview", func(url gson.JSON) (interface{}, error) {
			fmt.Printf("Attempting to generate preview for: %s\n", url.Str())
			select {
			case <-rateLimiter:
				p := pool.MustGet(create)
				defer pool.Put(p)

				done := make(chan string, 1)
				errCh := make(chan error, 1)

				go func() {
					defer func() {
						if r := recover(); r != nil {
							errCh <- fmt.Errorf("preview generation panic: %v", r)
						}
					}()

					fmt.Printf("Navigating to: %s\n", url.Str())
					p.MustNavigate(url.Str())

					// Add wait for page load
					p.MustWaitStable()

					fmt.Println("Taking screenshot...")
					screenshot := p.MustScreenshot()
					fmt.Printf("Screenshot taken, size: %d bytes\n", len(screenshot))
					done <- base64.StdEncoding.EncodeToString(screenshot)
				}()

				select {
				case result := <-done:
					fmt.Println("Preview generated successfully")
					return result, nil
				case err := <-errCh:
					fmt.Printf("Error generating preview: %v\n", err)
					return "", err
				case <-time.After(previewTimeout):
					fmt.Println("Preview generation timed out")
					return "", fmt.Errorf("preview generation timeout")
				}
			default:
				fmt.Println("Rate limit exceeded")
				return "", fmt.Errorf("rate limit exceeded")
			}
		})
	})()
}

func getJSCode() string {
	jsLib := loadLibraries()
	return fmt.Sprintf(`window.addEventListener('load', () => {
        console.log('Page loaded, initializing preview system');
        %s  // This loads the libraries

        // Initialize tippy globally
        tippy.setDefaultProps({
            trigger: 'manual',
            content: 'loading...',
            maxWidth: 500,
            interactive: true,
            placement: 'right',
        });

        function setup(el) {
            if (el.classList.contains('x-set')) return;
            console.log('Setting up preview for:', el.href);
            el.classList.add('x-set');

            const instance = tippy(el);

            // Custom show handler
            const showPreview = async () => {
                try {
                    console.log('Showing preview for:', el.href);
                    let img = document.createElement('img');
                    img.style.width = '400px';
                    console.log('Requesting preview...');
                    const preview = await getPreview(el.href);
                    console.log('Preview received:', preview ? 'success' : 'empty');
                    if (!preview) {
                        instance.setContent('Preview unavailable');
                        return;
                    }
                    img.src = "data:image/png;base64," + preview;
                    instance.setContent(img);
                } catch (err) {
                    console.error('Preview error:', err);
                    instance.setContent('Error: ' + err.message);
                }
            };

            // Handle right-click
            el.addEventListener('contextmenu', (e) => {
                e.preventDefault();
                instance.show();
                showPreview();
            });

            // Handle click anywhere else to hide
            document.addEventListener('click', (e) => {
                if (!el.contains(e.target)) {
                    instance.hide();
                }
            });
        }

        const observer = new MutationObserver((mutations) => {
            console.log('DOM mutation detected');
            mutations.forEach((mutation) => {
                mutation.addedNodes.forEach((node) => {
                    if (node.nodeType === 1) {
                        const links = node.matches('a') ? [node] :
                            Array.from(node.getElementsByTagName('a'));
                        links.forEach(setup);
                    }
                });
            });
        });

        console.log('Setting up initial links');
        document.querySelectorAll('a:not(.x-set)').forEach(setup);

        console.log('Starting MutationObserver');
        observer.observe(document.body, {
            childList: true,
            subtree: true
        });
    })`, jsLib)
}

func loadLibraries() string {
	// Make sure we load the libraries in the correct order
	libraries := []struct {
		name string
		url  string
	}{
		{"popper.js", "https://unpkg.com/@popperjs/core@2"},
		{"tippy.js", "https://unpkg.com/tippy.js@6"},
	}

	var combined string
	for _, lib := range libraries {
		combined += loadLibrary(lib.name, lib.url) + ";\n"
	}
	return combined
}

func loadLibrary(name, url string) string {
	cachePath := filepath.Join(cacheDir, name+".js")

	// Try to load from cache first
	if content, err := os.ReadFile(cachePath); err == nil {
		return string(content)
	}

	// Fetch and cache if not found
	res, err := http.Get(url)
	utils.E(err)
	defer res.Body.Close()

	content, err := io.ReadAll(res.Body)
	utils.E(err)

	// Save to cache
	err = os.WriteFile(cachePath, content, 0644)
	utils.E(err)

	return string(content)
}

func ensureCacheDir() {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		utils.E(err)
	}
}

func waitExit() {
	fmt.Println("Press Enter to exit...")
	utils.E(fmt.Scanln())
	os.Exit(0)
}
