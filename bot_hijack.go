package wee

import (
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// HijackHandler hijacks network resources, you can disable hijacked resources by setting the return value to true.
type HijackHandler func(*rod.Hijack)

// HijackAny
func (b *Bot) HijackAny(resources []string, terminate bool, handler HijackHandler) {
	if len(resources) == 0 {
		return
	}

	router := b.browser.HijackRequests()

	router.MustAdd("*", func(ctx *rod.Hijack) {
		for _, res := range resources {
			if strings.Contains(ctx.Request.URL().String(), res) {
				handler(ctx)

				if terminate {
					return
				}
			}
		}

		ctx.ContinueRequest(&proto.FetchContinueRequest{})
	})

	go router.Run()
}

func (b *Bot) Hijack(patterns []string, networkResourceType proto.NetworkResourceType, handler HijackHandler, continueRequest bool) {
	if len(patterns) == 0 {
		return
	}

	router := b.browser.HijackRequests()

	for _, pattern := range patterns {
		router.MustAdd(pattern, func(ctx *rod.Hijack) {
			if ctx.Request.Type() == networkResourceType {
				handler(ctx)

				if continueRequest {
					return
				}
			}

			ctx.ContinueRequest(&proto.FetchContinueRequest{})
		})
	}

	go router.Run()
}

func (b *Bot) DisableImages() {
	b.Hijack([]string{"*"},
		proto.NetworkResourceTypeImage,
		func(h *rod.Hijack) {
			h.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
		}, false)
}

func (b *Bot) DumpXHR(ptn []string, handler func(h *rod.Hijack)) {
	b.Hijack(ptn,
		proto.NetworkResourceTypeXHR,
		func(h *rod.Hijack) {
			h.MustLoadResponse()
			handler(h)
		}, true)
}
