package wee

import (
	"testing"

	"github.com/pterm/pterm"
	"github.com/stretchr/testify/suite"
)

type MiscSuite struct {
	suite.Suite
}

func TestMisc(t *testing.T) {
	suite.Run(t, new(MiscSuite))
}

func (s *MiscSuite) SetupSuite() {
}

func (s *MiscSuite) TearDownSuite() {
}

func (s *MiscSuite) Test_00_log() {
	pterm.PrintDebugMessages = true
	pterm.Debug.Printf("hello %s debug\n", "golang")
	pterm.Description.Printf("hello %s description\n", "golang")
	pterm.Info.Printf("hello %s info\n", "golang")
	pterm.Warning.Printf("hello %s warning\n", "golang")
	pterm.Error.Printf("hello %s error\n", "golang")
	// pterm.Fatal.Printf("hello %s fatal\n", "golang")
}

func (s *MiscSuite) Test_01_strToNum() {
	tests := []struct {
		raw   string
		kept  string
		wantS string
		wantF float64
	}{
		{"398 ofert pracy", "", "398", 398},
		{"398.1 ofert pracy", ".", "398.1", 398.1},
		{"ofert pracy", ".", "", 0},
		{"1,398.1 ofert pracy", ".", "1398.1", 1398.1},
	}
	for _, tt := range tests {
		v := StrToNumChars(tt.raw, tt.kept)
		s.Equal(tt.wantS, v)

		f := MustStrToFloat(tt.raw, tt.kept)
		s.Equal(tt.wantF, f)
	}
}
