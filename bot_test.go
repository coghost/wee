package wee

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type BotSuite struct {
	suite.Suite
}

func TestBot(t *testing.T) {
	suite.Run(t, new(BotSuite))
}

func (s *BotSuite) SetupSuite() {
}

func (s *BotSuite) TearDownSuite() {
}

func (s *BotSuite) Test_00_init() {
	w := NewBotWithDefault()
	w.MustOpen("http://www.baidu.com")
}
