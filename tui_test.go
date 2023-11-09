package wee

import (
	"testing"

	"github.com/pterm/pterm"
	"github.com/stretchr/testify/suite"
)

type TuiSuite struct {
	suite.Suite
}

func TestTui(t *testing.T) {
	suite.Run(t, new(TuiSuite))
}

func (s *TuiSuite) SetupSuite() {
}

func (s *TuiSuite) TearDownSuite() {
}

func (s *TuiSuite) Test_00_log() {
	pterm.PrintDebugMessages = true
	pterm.Debug.Printf("hello %s debug\n", "golang")
	pterm.Description.Printf("hello %s description\n", "golang")
	pterm.Info.Printf("hello %s info\n", "golang")
	pterm.Warning.Printf("hello %s warning\n", "golang")
	pterm.Error.Printf("hello %s error\n", "golang")
	pterm.Fatal.Printf("hello %s fatal\n", "golang")
}
