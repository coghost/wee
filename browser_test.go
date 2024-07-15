package wee

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type BrowserSuite struct {
	suite.Suite
}

func TestBrowser(t *testing.T) {
	suite.Run(t, new(BrowserSuite))
}

func (s *BrowserSuite) SetupSuite() {
}

func (s *BrowserSuite) TearDownSuite() {
}

func (s *BrowserSuite) TestNewLauncher() {
	client := NewLauncher(BrowserHeadless(true))
	s.Contains(client.MustLaunch(), "ws://127.0.0.1:")
}

func (s *BrowserSuite) TestNewBrowser() {
	exts := []string{"fixtures/chrome-extension"}
	dataDir := "/tmp/001"

	l, brw := NewBrowser(
		BrowserExtensions(exts),
		BrowserPaintRects(true),
		BrowserIncognito(true),
		BrowserUserDataDir(dataDir),
	)
	s.NotNil(l)
	s.NotNil(brw)

	defer l.Cleanup()
	defer brw.Close()

	ff, b := l.GetFlags("load-extension")
	s.True(b)
	s.Equal(exts, ff)

	v, b := l.GetFlags("user-data-dir")
	s.True(b)
	s.Contains(v, dataDir)

	brw.MustPage(fixtureFile("hellowee.html"))
	// uncomment Blocked() to check the extension manually.
	// Blocked()
	SleepN(2)
}
