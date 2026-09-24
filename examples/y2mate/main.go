package main

import (
	"log"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/coghost/wee"
)

var sites = []string{
	"https://cnvmp3.com/v54",
	"https://v3.y2mate.nu/",
	"https://yt2mp3.gs/",
	"",
	"",
}

// https://www.youtube.com/watch?v=h89XR-Es0II

const (
	AnotherSel  = `button[class="another"]`
	DownloadSel = `button[class=download]:not([disabled])`
	InputSel    = `input#video`
	submitSel   = `button[type=submit]`
)

func main() {
	bot := wee.NewBotDefault()
	defer bot.BlockInCleanUp()

	video := "https://www.youtube.com/watch?v=6T40TsMN08U"

	bot.MustOpen("https://v3.y2mate.nu/")

	got, err := bot.AnyElem([]string{
		AnotherSel, InputSel,
	}, wee.WithTimeout(wee.MediumToSec))

	if err != nil {
		log.Printf("load input/re-input has error: %v", err)
		return
	}

	if got == AnotherSel {
		bot.MustClick(got)
	}

	bot.MustInput(InputSel, video, wee.WithTimeout(wee.PT10Sec))
	bot.MustClick(submitSel)

	wee.SleepN(10)

	err = retry.Do(
		func() error {
			return bot.Click(DownloadSel, wee.WithTimeout(wee.MediumToSec))
		},
		retry.LastErrorOnly(false),
		retry.Attempts(5), //nolint:mnd
		retry.Delay(time.Second*1))
	if err != nil {
		log.Printf("cannot download mp3 %v", err)
	}

	wee.Confirm("please wait download complete and press enter to quit")
}
