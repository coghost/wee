package wee

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

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

func (s *BotSuite) Test00NewBot() {
	s.T().Parallel()
	want := "<html><head></head>\n<body>\nhellowee\n\n\n</body></html>"

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

			tt.c.MustOpen(fixtureFile("hellowee.html"))
			raw := tt.c.page.MustHTML()
			raw = strings.ReplaceAll(raw, " ", "")

			s.Equal(tt.wantS, raw, "")
		})
	}
}

func (s *BotSuite) Test00NewBotWithOptionsOnly() {
	bot := NewBotWithOptionsOnly()
	defer bot.Cleanup()

	s.Nil(bot.page)
	s.Panics(func() {
		bot.MustOpen(fixtureFile("helloworld.html"))
	}, "new bot only with options panics when open url")

	BindBotLanucher(bot)

	s.NotNil(bot.page)
	s.NotPanics(func() {
		bot.MustOpen(fixtureFile("helloworld.html"))
	}, "after binding bot with browser, shouldn't panic")
}

func (s *BotSuite) Test01CustomizePage() {
}
