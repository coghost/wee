package main

import (
	"fmt"

	"github.com/coghost/wee"
	"github.com/k0kubun/pp/v3"
)

// var botDetectionSites = []string{
// 	"https://bot.incolumitas.com/",
// 	"https://abrahamjuliot.github.io/creepjs/",
// 	"https://browserleaks.com/",
// 	"https://ipapi.is/",
// }

func main() {
	bot := wee.NewBotDefault()
	defer bot.BlockInCleanUp()

	var (
		home   = "https://bot.incolumitas.com/"
		inputs = []string{
			`div#formStuff input[name="userName"]`,
			`div#formStuff input[name="eMail"]`,
			`div#formStuff select`,
			`div#formStuff input[name="terms"]`,
			`div#formStuff input[id="%sCat"]`,
			`div#formStuff button`,
		}
	)

	bot.MustOpen(home)

	vals := []string{
		"John Reese",
		"john.r@poi.com",
		"I want all the Cookies",
	}

	bot.MustInput(inputs[0], vals[0], wee.WithHumanized(true))
	bot.MustInput(inputs[1], vals[1], wee.WithHumanized(true))

	bot.MustClick(inputs[2])
	err := bot.InputSelect(inputs[2], vals[2])
	if err != nil {
		panic(err)
	}

	bot.MustClick(inputs[3])
	bot.MustClick(fmt.Sprintf(inputs[4], "smol"))
	bot.MustClick(fmt.Sprintf(inputs[4], "big"))
	bot.MustClick(inputs[5])

	wait, handle := bot.Page().MustHandleDialog()
	wait()
	handle(true, "")

	wee.RandSleepNap()
	bot.MustAnyElem([]string{`#tableStuff tbody tr .url`})
	bot.MustClick(`div#tableStuff button#updatePrice0`)
	bot.MustClick(`div#tableStuff button#updatePrice1`)

	res := bot.MustEval(`() => {
	  let results = [];
	  document.querySelectorAll('#tableStuff tbody tr').forEach((row) => {
	    results.push({
	      name: row.querySelector('.name').innerText,
	      price: row.querySelector('.price').innerText,
	      url: row.querySelector('.url').innerText,
	    })
	  })
	  return results;}`)

	pp.Println(res)
}
