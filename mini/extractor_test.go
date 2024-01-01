package mini

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ExtractorSuite struct {
	suite.Suite
}

func TestExtractor(t *testing.T) {
	suite.Run(t, new(ExtractorSuite))
}

func (s *ExtractorSuite) SetupSuite() {
}

func (s *ExtractorSuite) TearDownSuite() {
}

func (s *ExtractorSuite) Test_01_strToNum() {
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

func (s *ExtractorSuite) Test_02_uniqid() {
	arr1 := []string{"a", "b"}
	arr2 := []string{"time", "now"}

	arr := append(arr1, arr2...)

	s.Equal([]string{"a", "b", "time", "now"}, arr)
}
