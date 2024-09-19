package wee

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coghost/wee/fixtures"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/stretchr/testify/suite"
)

type BotHijackSuite struct {
	suite.Suite
	ts  *httptest.Server
	bot *Bot
}

func TestBotHijack(t *testing.T) {
	suite.Run(t, new(BotHijackSuite))
}

func (s *BotHijackSuite) SetupSuite() {
	s.ts = fixtures.NewTestServer()
	s.bot = NewBotDefault()
}

func (s *BotHijackSuite) TearDownSuite() {
	s.bot.Cleanup()
	s.ts.Close()
}

func (s *BotHijackSuite) TestHijackAny() {
	intercepted := false
	resourceURL := ""

	// Set up the hijack before navigating
	s.bot.HijackAny([]string{"test-resource"}, func(h *rod.Hijack) {
		intercepted = true
		resourceURL = h.Request.URL().String()
		fmt.Printf("Intercepted request: %s\n", resourceURL)
		h.ContinueRequest(&proto.FetchContinueRequest{})
	}, false) // Set continueRequest to false to ensure our handler is responsible for continuing

	// Navigate to the page
	url := s.ts.URL + "/hijack_test?resource=test-resource"
	fmt.Printf("Navigating to: %s\n", url)
	s.bot.MustOpen(url)

	// Wait for a moment to allow for asynchronous interception
	time.Sleep(2 * time.Second)

	// Assert and provide detailed output
	s.True(intercepted, "HijackAny should have intercepted the request")
	s.Contains(resourceURL, "test-resource", "Intercepted URL should contain 'test-resource'")

	if !intercepted {
		fmt.Println("Interception failed. Current page URL:", s.bot.page.MustInfo().URL)
		// Optionally, print the page content for debugging
		fmt.Println("Page content:", s.bot.page.MustHTML())
	}
}

func (s *BotHijackSuite) TestHijack() {
	intercepted := false
	s.bot.Hijack([]string{"*.js"}, proto.NetworkResourceTypeScript, func(h *rod.Hijack) {
		intercepted = true
	}, true)

	s.bot.MustOpen(s.ts.URL + "/hijack_test")
	s.True(intercepted, "Hijack should have intercepted the JavaScript request")
}

func (s *BotHijackSuite) TestDisableImages() {
	s.bot.DisableImages()
	s.bot.MustOpen(s.ts.URL + "/image_test")

	// Check if images are actually disabled
	imageLoaded := s.bot.page.MustEval(`() => {
		const img = document.querySelector('img');
		return img.complete && img.naturalHeight !== 0;
	}`).Bool()

	s.False(imageLoaded, "Images should be disabled")
}

func (s *BotHijackSuite) TestDumpXHR() {
	xhrIntercepted := false
	var interceptedURL string
	var responseBody string

	// Use a pattern with prefix and suffix *
	s.bot.DumpXHR([]string{"*api/data*"}, func(h *rod.Hijack) {
		xhrIntercepted = true
		interceptedURL = h.Request.URL().String()
		responseBody = h.Response.Body()
		fmt.Printf("Intercepted XHR: %s\n", interceptedURL)
		fmt.Printf("Response body: %s\n", responseBody)
	})

	// Navigate to the page
	url := s.ts.URL + "/xhr_test"
	fmt.Printf("Navigating to: %s\n", url)
	s.bot.MustOpen(url)

	// Wait for the XHR to complete
	s.bot.page.MustWaitRequestIdle()

	// Add a small delay to ensure the handler has time to execute
	time.Sleep(500 * time.Millisecond)

	s.True(xhrIntercepted, "DumpXHR should have intercepted the XHR request")
	s.Contains(interceptedURL, "api/data", "Intercepted URL should contain 'api/data'")
	s.Contains(responseBody, "XHR response", "Response body should contain expected content")

	if !xhrIntercepted {
		fmt.Println("XHR interception failed. Current page URL:", s.bot.page.MustInfo().URL)
		fmt.Println("Page content:", s.bot.page.MustHTML())
	}
}

func (s *BotHijackSuite) TestHijackAnyWithMultipleResources() {
	interceptCount := 0
	s.bot.HijackAny([]string{"resource1", "resource2"}, func(h *rod.Hijack) {
		interceptCount++
	}, true)

	s.bot.MustOpen(s.ts.URL + "/hijack_test?resource=resource1")
	s.bot.MustOpen(s.ts.URL + "/hijack_test?resource=resource2")
	s.GreaterOrEqual(interceptCount, 2, "HijackAny should have intercepted both resources")
}

func (s *BotHijackSuite) TestHijackWithMultiplePatterns() {
	interceptCount := 0
	s.bot.Hijack([]string{"*.js", "*.css"}, proto.NetworkResourceTypeStylesheet, func(h *rod.Hijack) {
		interceptCount++
	}, true)

	s.bot.MustOpen(s.ts.URL + "/hijack_test_multiple")
	s.GreaterOrEqual(interceptCount, 1, "Hijack should have intercepted at least one stylesheet")
}
