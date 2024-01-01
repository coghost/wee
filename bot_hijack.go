package wee

import (
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/rs/zerolog/log"
)

type HijactResp struct {
	Pattern string
	Uri     string
	Body    string
}

func (b *Bot) HijackNetwork(brw *rod.Browser, patterns []string, resourceType proto.NetworkResourceType, callback func(a, b string)) {
	if len(patterns) == 0 {
		return
	}

	router := brw.HijackRequests()

	for _, pattern := range patterns {
		router.MustAdd(pattern, func(ctx *rod.Hijack) {
			if ctx.Request.Type() == resourceType {
				ctx.MustLoadResponse()
				body := ctx.Response.Body()
				uri := ctx.Request.URL().String()
				log.Info().Str("pattern", pattern).Str("uri", uri).Msg("hijacked")
				callback(uri, body)
			}
			ctx.ContinueRequest(&proto.FetchContinueRequest{})
		})
	}

	go router.Run()
}

func (b *Bot) HijackDocments(brw *rod.Browser, patterns []string, callback func(a, b string)) {
	b.HijackNetwork(brw, patterns, proto.NetworkResourceTypeDocument, callback)
}
