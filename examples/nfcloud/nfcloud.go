package main

import (
	"github.com/coghost/wee"
	"github.com/coghost/xpretty"
)

const (
	home     = "https://nfcloud.net/auth/login"
	login    = `input#email`
	username = `ul.navbar-nav span.avatar-sm+div`
)

func main() {
	bot := wee.NewBotDefault(
		wee.WithCookies(true),
	)

	defer bot.Cleanup()
	defer wee.Blocked()

	bot.MustOpen(home)

	whom := bot.MustAnyElem(login, username)
	if whom == login {
		wee.Confirm("sign in and continue")
		bot.DumpCookies()
	}

	name := bot.MustElemAttr(username)
	xpretty.CyanPrintf("got username %s\n", name)
}
