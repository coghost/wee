package wee

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type LaunchSuite struct {
	suite.Suite
}

func TestLaunch(t *testing.T) {
	suite.Run(t, new(LaunchSuite))
}

func (s *LaunchSuite) SetupSuite() {
}

func (s *LaunchSuite) TearDownSuite() {
}

func (s *LaunchSuite) Test_00_new() {
	l := NewLauncher()
	s.Contains(l.MustLaunch(), "ws://127.0.0.1:")
}
