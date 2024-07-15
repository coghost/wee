package wee

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/ungerik/go-dry"
)

// fixtures serves fixtures/xxx as `file:///xxx`
func fixtures(path string) string {
	slash := filepath.FromSlash
	f, err := filepath.Abs(slash("fixtures/" + path))
	dry.PanicIfErr(err)

	return "file://" + f
}

type BotSuite struct {
	suite.Suite
	// ts *httptest.Server
}

func TestBot(t *testing.T) {
	suite.Run(t, new(BotSuite))
}

func (s *BotSuite) SetupSuite() {
	// s.ts = newTestServer()
}

func (s *BotSuite) TearDownSuite() {
}

func (s *BotSuite) Test00newBot() {
	s.T().Parallel()
	want := "<html><head></head>\n<body>\nhelloworld\n\n\n</body></html>"

	tests := []struct {
		name  string
		c     *Bot
		wantS string
	}{
		{
			name:  "default_bot",
			c:     NewBotDefault(),
			wantS: want,
		},
		{
			name:  "debug_bot",
			c:     NewBotForDebug(),
			wantS: want,
		},
		{
			name:  "headless_bot",
			c:     NewBot(Headless(true)),
			wantS: want,
		},
	}

	for _, tt := range tests {
		s.T().Run("test_"+tt.name, func(t *testing.T) {
			t.Parallel()

			tt.c.MustOpen(fixtures("helloworld.html"))
			raw := tt.c.page.MustHTML()
			raw = strings.ReplaceAll(raw, " ", "")

			s.Equal(tt.wantS, raw, "")
		})
	}
}

func (s *BotSuite) Test00newBotWithOptionsOnly() {
	s.T().Parallel()

	c := NewBotWithOptionsOnly()
	s.Panics(func() {
		c.MustOpen(fixtures("helloworld.html"))
	}, "new bot only with options panics when open url")

	BindBotLanucher(c)
	s.NotPanics(func() {
		c.MustOpen(fixtures("helloworld.html"))
	}, "after binding bot with browser, shouldn't panic")
}
