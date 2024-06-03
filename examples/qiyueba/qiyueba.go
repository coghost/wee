package main

import (
	"fmt"

	"github.com/coghost/wee"
	"github.com/coghost/xlog"
	"github.com/coghost/xpretty"
)

const (
	home     = "https://www.qiyueba.com/"
	input    = `input[name="s"]`
	submit   = `button.btn`
	filter   = `a[href="#%s"]`
	username = `a[href="/wp-admin"]`
	login    = `a[href="/wp-login.php"]`
)

const (
	_newposts = "newposts"
	_dayhot   = "dayhot"
	_weekhot  = "weekhot"
)

func main() {
	xlog.InitLogForConsole()

	bot := wee.NewBotDefault(
		wee.WithCookieFile("/tmp/001.json"),
	)

	defer bot.Cleanup()
	defer wee.Blocked()

	bot.MustOpen(home)

	got := bot.MustAnyElem(login, username)

	if got == username {
		name := bot.MustElemAttr(username)
		xpretty.CyanPrintf("got username %s\n", name)

		bot.MustClick(fmt.Sprintf(filter, _dayhot), wee.WithWaitStable(true))
		bot.DumpCookies()
	} else {
		if b := wee.Confirm("sign in manually and quit"); b {
			bot.DumpCookies()
		}
	}
}
