package wee

import (
	"fmt"
	"os"
	"time"

	"github.com/coghost/xdtm"
	"github.com/coghost/xpretty"
	"github.com/pterm/pterm"
)

func SleepWithSpin(n int, args ...string) {
	msg := fmt.Sprintf("Wait for %d seconds before quitting ...", n)
	if len(args) > 0 {
		msg = args[0]
	}

	spinnerInfo, _ := pterm.DefaultSpinner.Start(xpretty.Yellow(msg))

	time.Sleep(time.Second * time.Duration(n))
	spinnerInfo.Info()
}

func Confirm(msg ...string) bool {
	b, _ := pterm.DefaultInteractiveConfirm.Show(msg...)
	return b
}

func QuitIfY() {
	pd := pterm.DefaultInteractiveConfirm
	pd.DefaultValue = false
	msg := fmt.Sprintf("%s: Wanna quit(Y) or continue(N)?", xdtm.CarbonNow().ToShortDateString())

	if b, _ := pd.Show(msg); b {
		os.Exit(0)
	}
}
