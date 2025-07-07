package wee

import (
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// HijackHandler is a function type used for handling hijacked network resources in web automation tasks.
//
// Type Definition:
//
//	type HijackHandler func(*rod.Hijack)
//
// Purpose:
//
//	This type defines a callback function that is invoked when a network request
//	matching specified criteria is intercepted during web automation or scraping tasks.
//
// Parameters:
//   - *rod.Hijack: A pointer to the rod.Hijack object, which encapsulates the
//     intercepted request and provides methods to inspect and manipulate both
//     the request and its response.
//
// Functionality:
//  1. Request Inspection: Examine details of the intercepted request such as URL,
//     headers, method, and body.
//  2. Response Manipulation: Modify the response before it reaches the browser,
//     including changing status codes, headers, or body content.
//  3. Request Blocking: Prevent certain requests from being sent.
//  4. Custom Logic: Implement any custom behavior based on the intercepted request.
//
// Usage Context:
//   - Used as a parameter in methods like HijackAny and Hijack of the Bot struct.
//   - Allows for fine-grained control over network traffic during web automation.
//
// Common Use Cases:
//  1. Modifying API responses for testing purposes.
//  2. Blocking unwanted resources (e.g., advertisements).
//  3. Logging or analyzing network traffic.
//  4. Injecting custom JavaScript or CSS into responses.
//
// Example Implementation:
//
//	handler := func(h *rod.Hijack) {
//	    // Log the URL of each request
//	    fmt.Println("Intercepted request to:", h.Request.URL().String())
//
//	    // Modify a specific API response
//	    if strings.Contains(h.Request.URL().String(), "api/data") {
//	        h.Response.SetBody([]byte(`{"modified":"data"}`))
//	    }
//
//	    // Block image requests
//	    if h.Request.Type() == proto.NetworkResourceTypeImage {
//	        h.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
//	    }
//	}
//
// Note:
//   - The exact behavior after the handler executes depends on the 'continueRequest'
//     parameter in the hijacking method and any actions taken within the handler.
//   - Care should be taken when modifying responses, as it may affect the functionality
//     of the web page being automated.
//
// Related Methods:
//   - Bot.HijackAny: Hijacks requests matching any of the specified resource strings.
//   - Bot.Hijack: Hijacks requests matching specified patterns and resource types.
//
// See Also:
//   - rod.Hijack documentation for more details on available methods and properties.
type HijackHandler func(*rod.Hijack)

// HijackAny intercepts network requests that contain any of the specified resource strings in their URLs.
//
//	This method sets up a network request interceptor for the bot's browser, allowing
//	inspection and manipulation of requests based on URL content.
//
// Parameters:
//
//  1. resources []string
//     - A slice of strings to match against request URLs.
//     - If a request URL contains any of these strings, it will be intercepted.
//     - Case-sensitive matching is used.
//     - If empty, the method returns without setting up interception.
//
//  2. handler HijackHandler
//     - A function of type func(*rod.Hijack) that processes intercepted requests.
//     - Called for each request that matches the resources criteria.
//     - Can inspect request details and modify the response.
//
//  3. continueRequest bool
//     - Determines the flow after handler execution.
//     - If true: The request continues its normal flow after handler execution.
//     - If false: The request is terminated after handler execution, unless explicitly continued in the handler.
//
// Use Cases:
//   - Monitoring specific API calls or resource loads.
//   - Modifying responses for testing or mocking purposes.
//   - Blocking unwanted resources (e.g., ads, trackers).
//   - Logging network activity for debugging.
//
// Example:
//
//	bot.HijackAny(
//	    []string{"api/users", "images"},
//	    func(h *rod.Hijack) {
//	        fmt.Printf("Intercepted: %s\n", h.Request.URL().String())
//	        if strings.Contains(h.Request.URL().String(), "api/users") {
//	            // Modify API response
//	            h.Response.SetBody([]byte(`{"users":[]}`))
//	        }
//	    },
//	    true
//	)
//
// Notes:
//   - This method starts a goroutine, so it returns immediately.
//   - Be cautious with resource-intensive operations in the handler to avoid performance issues.
//   - Modifying responses can affect page functionality; use with care.
//
// See Also:
//   - Bot.Hijack: For more specific request targeting by pattern and resource type.
//   - rod.Hijack documentation for available methods on the Hijack object.
func (b *Bot) HijackAny(resources []string, handler HijackHandler, continueRequest bool) {
	if len(resources) == 0 {
		return
	}

	router := b.browser.HijackRequests()

	router.MustAdd("*", func(ctx *rod.Hijack) {
		for _, res := range resources {
			if strings.Contains(ctx.Request.URL().String(), res) {
				handler(ctx)

				if !continueRequest {
					return
				}
			}
		}

		ctx.ContinueRequest(&proto.FetchContinueRequest{})
	})

	go router.Run()
}

// Hijack intercepts network requests that match specified patterns and resource types.
//
//	This method sets up a network request interceptor for the bot's browser, allowing
//	inspection and manipulation of requests based on URL patterns and resource types.
//
// Parameters:
//
//  1. patterns []string
//     - A slice of URL patterns to match against request URLs.
//     - Supports glob patterns (e.g., "*.example.com").
//     - If empty, the method returns without setting up interception.
//
//  2. networkResourceType proto.NetworkResourceType
//     - Specifies the type of network resource to intercept (e.g., XHR, Image, Script).
//     - Only requests of this type will be intercepted.
//
//  3. handler HijackHandler
//     - A function of type func(*rod.Hijack) that processes intercepted requests.
//     - Called for each request that matches both the pattern and resource type criteria.
//
//  4. continueRequest bool
//     - Determines whether to continue the request after handler execution.
//     - If true: The request continues its normal flow after handler execution.
//     - If false: The request is terminated after handler execution, unless explicitly continued in the handler.
//
// Implementation Details:
//   - Uses rod's pattern matching for URLs, which supports glob patterns.
//   - Only requests matching both pattern and resource type are intercepted.
//   - Non-matching requests (by type) are automatically continued.
//
// Use Cases:
//   - Intercepting specific types of requests (e.g., only XHR or Image requests).
//   - Modifying responses for certain domains or URL patterns.
//   - Implementing fine-grained control over network traffic during web automation.
//
// Example:
//
//	bot.Hijack(
//	    []string{"api.example.com/*"},
//	    proto.NetworkResourceTypeXHR,
//	    func(h *rod.Hijack) {
//	        fmt.Printf("Intercepted XHR: %s\n", h.Request.URL().String())
//	        // Modify XHR response
//	        h.Response.SetBody([]byte(`{"status":"mocked"}`))
//	    },
//	    true
//	)
//
// Notes:
//   - This method starts a goroutine, so it returns immediately.
//   - Be cautious with resource-intensive operations in the handler to avoid performance issues.
//   - Modifying responses can affect page functionality; use with care.
//
// See Also:
//   - Bot.HijackAny: For intercepting requests based on URL content without type restrictions.
//   - proto.NetworkResourceType documentation for available resource types.
//   - rod.Hijack documentation for available methods on the Hijack object.
func (b *Bot) Hijack(patterns []string, networkResourceType proto.NetworkResourceType, handler HijackHandler, continueRequest bool) {
	if len(patterns) == 0 {
		return
	}

	router := b.browser.HijackRequests()

	for _, pattern := range patterns {
		router.MustAdd(pattern, func(ctx *rod.Hijack) {
			if ctx.Request.Type() == networkResourceType {
				handler(ctx)

				if !continueRequest {
					return
				}
			}

			ctx.ContinueRequest(&proto.FetchContinueRequest{})
		})
	}

	go router.Run()
}

// DisableImages blocks all image requests for the bot's browser session.
//
// This method sets up a request interceptor that fails all requests for image resources,
// effectively preventing images from loading during web navigation.
//
// Usage:
//
//	bot.DisableImages()
//
// Note:
//   - This method is useful for reducing bandwidth usage and speeding up page loads.
//   - It affects all subsequent page loads until the bot is reset or the browser is closed.
func (b *Bot) DisableImages() {
	b.Hijack([]string{"*"},
		proto.NetworkResourceTypeImage,
		func(h *rod.Hijack) {
			h.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
		}, false)
}

// DumpXHR intercepts and processes XMLHttpRequest (XHR) requests matching specified patterns.
//
// Parameters:
//   - ptn []string: URL patterns to match XHR requests. Use "*" as prefix and suffix for partial matches.
//   - handler func(h *rod.Hijack): Function to process each intercepted XHR.
//
// The method loads the full response for each matched XHR and calls the handler,
// allowing inspection of both request and response details without blocking the XHR.
//
// Usage:
//
//	bot.DumpXHR([]string{"*api/data*"}, func(h *rod.Hijack) {
//	    h.MustLoadResponse()
//	    fmt.Printf("XHR: %s\n", h.Request.URL())
//	    fmt.Printf("Response Body: %s\n", h.Response.Body())
//	})
//
// Pattern examples:
//   - "*api*": Matches any URL containing "api"
//   - "*example.com/api*": Matches URLs containing "example.com/api"
//   - "*": Matches all XHR requests
//
// Note:
//   - Patterns are case-sensitive.
//   - The handler is called after the response is loaded, allowing access to both request and response data.
//   - This method is useful for logging, debugging, or analyzing specific XHR traffic during web automation.
//
// Be cautious:
//   - Overly broad patterns may intercept more requests than intended.
//   - Intercepting many XHRs may impact performance. Use specific patterns when possible.
//   - Modifying XHR responses can affect the functionality of the web application being automated.
func (b *Bot) DumpXHR(ptn []string, handler func(h *rod.Hijack)) {
	b.Hijack(ptn,
		proto.NetworkResourceTypeXHR,
		func(h *rod.Hijack) {
			// h.MustLoadResponse()
			handler(h)
		}, true)
}
