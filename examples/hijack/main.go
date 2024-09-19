package main

import (
	"log"
	"net/http"
	"net/url"

	"github.com/coghost/wee"
	"github.com/coghost/zlog"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"go.uber.org/zap"
)

func main() {
	logger := zlog.MustNewZapLogger()
	ua := `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36`

	bot := wee.NewBotWithOptionsOnly(
		wee.Headless(false),
		wee.UserAgent(ua),
		wee.AcceptLanguage("zh_CN"),
		wee.WithCookies(false),
		wee.Logger(logger),
	)
	defer bot.Cleanup()

	wait := make(chan byte)

	wee.BindBotLanucher(bot)

	bot.Hijack([]string{
		"*getPositionList",
	}, proto.NetworkResourceTypeXHR, func(ctx *rod.Hijack) {
		body := ctx.Request.JSONBody()
		logger.Info("Request body", zap.Any("body", body.String()))
		_ = ctx.LoadResponse(http.DefaultClient, true)
		raw := ctx.Response.Body()
		logger.Info("Response body", zap.String("body", raw))
	}, true)

	bot.MustOpen("https://talent.pingan.com/recruit/social.html")
	bot.Page().MustWaitDOMStable()
	logger.Info("DOM is ready")

	<-wait
}

// pageDisableCache
// https://github.com/go-rod/rod/issues/1063
// https://github.com/go-rod/rod/issues/851
func pageDisableCache(page *rod.Page) {
	router := page.HijackRequests()
	err := router.Add("", "", func(ctx *rod.Hijack) {
		ul, _ := url.Parse("socks5://127.0.0.1:8888")
		proxy := http.ProxyURL(ul)
		transport := &http.Transport{Proxy: proxy}
		error := ctx.LoadResponse(&http.Client{Transport: transport}, true)
		if error != nil {
			log.Println("Hijack LoadResponse err:", error)
			return
		}
		ctx.Response.SetHeader("Cache-Control", "no-cache, no-store, must-revalidate")
		ctx.Response.SetHeader("Pragma", "no-cache")
		ctx.Response.SetHeader("Expires", "0")
	})
	if err != nil {
		log.Println("Hijack response headers err:", err)
		return
	}
	go router.Run()
}
