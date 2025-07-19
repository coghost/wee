package main

import (
	"fmt"
	"log"

	"github.com/coghost/pathlib"
	"github.com/coghost/wee"
)

type Subtitle struct {
	Lang     string `json:"lang,omitempty"`
	LangDesc string `json:"langDesc,omitempty"`
	Content  string `json:"content,omitempty"`
}

type SubtitleInfo struct {
	Vid       string `json:"vid,omitempty"`
	Host      string `json:"host,omitempty"`
	HostAlias string `json:"hostAlias,omitempty"`
	Title     string `json:"title,omitempty"`
	Status    string `json:"status,omitempty"`

	Subtitles []Subtitle `json:"subtitleItemVoList,omitempty"`
}

const (
	url    = `https://www.kedou.life/caption/subtitle/bilibili`
	input  = `input.el-input__inner`
	submit = `button.el-button`
	items  = `a@@@下载`

	script = `() => { return JSON.stringify(window.__NUXT__.pinia['captionStore']['subtitleExtractInfo']) }`
)

func main() {
	keyword := `https://www.bilibili.com/video/BV1J5XZYPEVa/?p=3`
	run(keyword)
}

func run(videoURL string) {
	bot := wee.NewBotDefault(wee.WithBounds(wee.NewBoundsHD()))

	bot.MustOpen(url)
	bot.MustInput(input, videoURL)
	bot.MustClick(submit)

	_, err := bot.Elems(items, wee.WithTimeout(wee.MediumToSec))
	if err != nil {
		log.Printf("cannot get video subtitles: %v", err)
	}

	rawRes := bot.MustEval(script)
	if err := pathlib.Path("/tmp/abc.json").SetString(rawRes); err != nil {
		log.Printf("cannot save %v", err)
	}
}

func parse() {
	var subInfo *SubtitleInfo

	err := pathlib.Path("/tmp/abc.json").GetJSON(&subInfo)
	if err != nil {
		log.Printf("%v", err)
	}

	// pp.Println(subInfo.Subtitles)

	title := subInfo.Title
	rootFs := pathlib.Path("/tmp/bilibili")

	fileFs := rootFs.Join(title)
	stem := fileFs.Stem
	suf := fileFs.Suffix

	for i, sub := range subInfo.Subtitles {
		langSrt := fmt.Sprintf("%s.%s.%d%s", stem, sub.Lang, i, suf)
		log.Printf("writing %s", langSrt)
		rootFs.Join(langSrt).SetString(sub.Content)
	}
}
