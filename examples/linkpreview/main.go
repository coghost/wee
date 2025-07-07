package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/rod/lib/utils"
	"github.com/ysmood/gson"
)

const (
	maxConcurrentScans = 2
	maxRequired        = 10
	scanTimeout        = 10 * time.Second
	outputDir          = "/tmp/scanned_pages" // Directory to store HTML files
)

// Modify main to create output directory
func main() {
	if err := ensureOutputDir(); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	wsURL := launcher.NewUserMode().MustLaunch()
	browser := rod.New().ControlURL(wsURL).MustConnect().NoDefaultDevice()

	setupScanner(browser)

	page := browser.MustPage()
	page.MustNavigate("https://httpbin.org/")

	waitExit()
}

// Add this function to sanitize URLs for filenames
func sanitizeFileName(url string) string {
	// Replace special characters with underscore
	invalid := []string{":", "/", "\\", "?", "*", "\"", "<", ">", "|", " "}
	name := url
	for _, char := range invalid {
		name = strings.ReplaceAll(name, char, "_")
	}
	return name + ".html"
}

// Add this function to ensure output directory exists
func ensureOutputDir() error {
	return os.MkdirAll(outputDir, 0755)
}

// Add this function to save HTML content
func saveHTML(url, html string) error {
	fileName := sanitizeFileName(url)
	filePath := filepath.Join(outputDir, fileName)
	return os.WriteFile(filePath, []byte(html), 0644)
}

func setupScanner(browser *rod.Browser) {
	fmt.Println("Initializing link scanner...")
	scanner := rod.New().MustConnect()
	scanner.MustSetCookies(browser.MustGetCookies()...)
	pool := rod.NewPagePool(maxConcurrentScans)
	create := func() *rod.Page { return scanner.MustPage() }

	go browser.EachEvent(func(e *proto.TargetTargetCreated) {
		if e.TargetInfo.Type != proto.TargetTargetInfoTypePage {
			return
		}
		fmt.Println("Setting up scanner for new page...")
		page := browser.MustPageFromTargetID(e.TargetInfo.TargetID)
		page.MustEvalOnNewDocument(getJSCode())

		page.MustExpose("scanPageLinks", func(urls gson.JSON) (interface{}, error) {
			urlList := urls.Arr()
			if maxRequired > 0 && len(urlList) > maxRequired {
				urlList = urlList[:maxRequired]
			}

			limitMsg := ""
			if maxRequired > 0 {
				limitMsg = fmt.Sprintf(" (limited to %d)", maxRequired)
			}
			fmt.Printf("Received scan request for %d URLs%s\n", len(urlList), limitMsg)

			results := make(map[string]string)
			sem := make(chan struct{}, maxConcurrentScans)
			errCount := 0
			successCount := 0

			for _, url := range urlList {
				sem <- struct{}{} // Acquire

				go func(urlStr string) {
					defer func() { <-sem }() // Release

					p := pool.MustGet(create)
					defer pool.Put(p)

					done := make(chan string, 1)
					errCh := make(chan error, 1)

					go func() {
						defer func() {
							if r := recover(); r != nil {
								errCh <- fmt.Errorf("scan panic: %v", r)
							}
						}()

						fmt.Printf("🔍 Scanning (%d/%d): %s\n", successCount+1, len(urlList), urlStr)
						p.MustNavigate(urlStr)
						p.MustWaitStable()

						html := p.MustElement("html").MustHTML()
						done <- html
					}()

					select {
					case html := <-done:
						results[urlStr] = html
						successCount++
						if err := saveHTML(urlStr, html); err != nil {
							fmt.Printf("❌ Error saving HTML for %s: %v\n", urlStr, err)
						} else {
							fmt.Printf("✅ Scanned and saved (%d/%d): %s\n", successCount, len(urlList), urlStr)
						}
					case err := <-errCh:
						errCount++
						fmt.Printf("❌ Error (%d errors) scanning %s: %v\n", errCount, urlStr, err)
					case <-time.After(scanTimeout):
						errCount++
						fmt.Printf("⏰ Timeout (%d errors) scanning: %s\n", errCount, urlStr)
					}
				}(url.Str())
			}

			// Wait for all scans to complete
			for i := 0; i < maxConcurrentScans; i++ {
				sem <- struct{}{}
			}

			completionMsg := fmt.Sprintf("📊 Scan complete: %d successful, %d failed",
				successCount, errCount)
			if maxRequired > 0 {
				completionMsg += fmt.Sprintf(" (max: %d)", maxRequired)
			}
			fmt.Println(completionMsg)

			return results, nil
		})
	})()
	fmt.Println("Link scanner initialized successfully")
}

func getJSCode() string {
	var sliceCode string
	if maxRequired > 0 {
		sliceCode = fmt.Sprintf(".slice(0, %d)", maxRequired)
	}

	var limitMsg string
	if maxRequired > 0 {
		limitMsg = fmt.Sprintf(" (limited to %d)", maxRequired)
	}

	return fmt.Sprintf(`
        let clickCount = 0;
        let clickTimer = null;

        document.addEventListener('click', async (e) => {
            clickCount++;

            if (clickTimer) {
                clearTimeout(clickTimer);
            }

            clickTimer = setTimeout(() => {
                if (clickCount === 3) {
                    console.log('Triple click detected, scanning links...');

                    // Save current page first
                    try {
                        const currentHTML = document.documentElement.outerHTML;
                        const currentURL = window.location.href;
                        window.scanPageLinks([currentURL]);
                        console.log('Current page saved');
                    } catch (err) {
                        console.error('Failed to save current page:', err);
                    }

                    // Then scan other links
                    const links = Array.from(document.getElementsByTagName('a'))
                        .map(a => a.href)
                        .filter(href => href && href.startsWith('http'))
                        .filter(href => href !== window.location.href)%s; // Exclude current page

                    if (links.length === 0) {
                        console.log('No additional links found to scan');
                        return;
                    }

                    console.log('Found ' + links.length + ' additional links to scan%s');
                    try {
                        window.scanPageLinks(links);
                        console.log('Scan initiated');
                    } catch (err) {
                        console.error('Scan failed', err);
                    }
                }
                clickCount = 0;
            }, 300); // 300ms window for triple click
        });
    `, sliceCode, limitMsg)
}

func waitExit() {
	fmt.Println("Press Enter to exit...")
	utils.E(fmt.Scanln())
	os.Exit(0)
}
