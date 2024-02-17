package mini

import (
	"fmt"
	"testing"

	"github.com/icza/gox/fmtx"
	"github.com/stretchr/testify/suite"
)

type ReqSuite struct {
	suite.Suite
}

func TestReq(t *testing.T) {
	suite.Run(t, new(ReqSuite))
}

func (s *ReqSuite) SetupSuite() {
}

func (s *ReqSuite) TearDownSuite() {
}

func (s *ReqSuite) Test_00_condSprintf() {
	type (
		_sprinter func(format string, a ...any) string
	)

	tests := []struct {
		name   string
		sprint _sprinter
		sel    string
		val    string
		wantS  string
	}{
		{"fmt.Sprintf", fmt.Sprintf, `div[style*=primary]`, "title", "div[style*=primary]%!(EXTRA string=title)"},
		{"fmt.Sprintf", fmt.Sprintf, `div[style*=primary]>div.[name=%s]`, "title", "div[style*=primary]>div.[name=title]"},
		// cond sprintf
		{"CondSprintf", fmtx.CondSprintf, `div[style*=primary]`, "title", "div[style*=primary]"},
		{"fmt.Sprintf", fmtx.CondSprintf, `div[style*=primary]>div.[name=%s]`, "title", "div[style*=primary]>div.[name=title]"},
	}
	for _, tt := range tests {
		got := tt.sprint(tt.sel, tt.val)
		s.Equal(tt.wantS, got, tt.name)
	}
}
