package wee

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/coghost/wee/fixtures"
	"github.com/go-rod/rod/lib/input"
	"github.com/stretchr/testify/suite"
)

type BotClickSuite struct {
	suite.Suite
	ts  *httptest.Server
	bot *Bot
}

func TestBotClick(t *testing.T) {
	suite.Run(t, new(BotClickSuite))
}

func (s *BotClickSuite) SetupSuite() {
	s.ts = fixtures.NewTestServer()
}

func (s *BotClickSuite) TearDownSuite() {
}

func (s *BotClickSuite) SetupTest() {
	s.bot = NewBotDefault()
}

func (s *BotClickSuite) TearDownTest() {
	s.bot.Cleanup()
}

func (s *BotClickSuite) TestMustClickAndWait() {
	s.T().Parallel()
	s.bot.MustOpen(s.ts.URL + "/click_test")
	s.bot.MustClickAndWait("#clickme")
	s.True(s.bot.page.MustHas("#clicked"))
}

func (s *BotClickSuite) TestMustClickAll() {
	s.T().Parallel()
	s.bot.MustOpen(s.ts.URL + "/click_test")
	s.bot.MustClickAll([]string{"#button1", "#button2", "#button3"})
	SleepN(1)
	s.True(s.bot.page.MustHas("#all-clicked"))
}

func (s *BotClickSuite) TestClick() {
	s.T().Parallel()
	s.bot.MustOpen(s.ts.URL + "/click_test")
	err := s.bot.Click("#clickme")
	s.Nil(err)
	s.True(s.bot.page.MustHas("#clicked"))

	err = s.bot.Click("#non-existent")
	s.NotNil(err)
}

func (s *BotClickSuite) TestClickElem() {
	s.T().Parallel()
	s.bot.MustOpen(s.ts.URL + "/click_test")
	elem := s.bot.page.MustElement("#clickme")
	err := s.bot.ClickElem(elem)
	s.Nil(err)
	s.True(s.bot.page.MustHas("#clicked"))
}

func (s *BotClickSuite) TestMakeElemInteractable() {
	s.T().Parallel()
	s.bot.MustOpen(s.ts.URL + "/click_test")
	elem := s.bot.page.MustElement("#hidden-button")
	err := s.bot.MakeElemInteractable(elem, true)
	s.NotNil(err)
}

func (s *BotClickSuite) TestInteractable() {
	s.T().Parallel()
	s.bot.MustOpen(s.ts.URL + "/click_test")
	elem := s.bot.page.MustElement("#offscreen-button")
	err := s.bot.Interactable(elem)
	s.NotNil(err)
}

func (s *BotClickSuite) TestPressEscape() {
	s.T().Parallel()
	s.bot.MustOpen(s.ts.URL + "/click_test")
	elem := s.bot.page.MustElement("body")
	err := s.bot.PressEscape(elem)
	s.Nil(err)
}

func (s *BotClickSuite) TestPress() {
	s.T().Parallel()
	s.bot.MustOpen(s.ts.URL + "/click_test")
	elem := s.bot.page.MustElement("#input-field")
	err := s.bot.Press(elem, input.Enter)
	s.Nil(err)
	s.True(s.bot.page.MustHas("#input-submitted"))
}

func (s *BotClickSuite) TestClickElemWithScript() {
	s.T().Parallel()
	s.bot.MustOpen(s.ts.URL + "/click_test")
	elem := s.bot.page.MustElement("#clickme")
	err := s.bot.ClickElemWithScript(elem)
	s.Nil(err)
	s.True(s.bot.page.MustHas("#clicked"))
}

func (s *BotClickSuite) TestAcceptCookies() {
	s.T().Parallel()
	s.bot.MustOpen(s.ts.URL + "/cookie_test")
	err := s.bot.AcceptCookies("#accept-cookies", "#accept-all")
	s.Nil(err)
	s.True(s.bot.page.MustHas("#cookies-accepted"))
}

func (s *BotClickSuite) TestClosePopovers() {
	s.T().Parallel()
	s.bot.MustOpen(s.ts.URL + "/popover_test")

	// Check that popovers are initially visible
	popovers := s.bot.page.MustElements(".popover")
	s.Len(popovers, 2)
	for _, popover := range popovers {
		display, err := popover.Eval(`() => this.style.display`)
		s.Nil(err)
		s.NotEqual("none", display.Value.String(), "Popover should be visible initially")
	}

	// Close popovers
	hit := s.bot.ClosePopovers(".popover-close")
	s.Equal(2, hit)

	// Check that popovers are now hidden
	for _, popover := range popovers {
		display, err := popover.Eval(`() => this.style.display`)
		s.Nil(err)
		s.Equal("none", display.Value.String(), "Popover should be hidden after closing")
	}
}

func (s *BotClickSuite) TestClosePopover() {
	s.T().Parallel()
	s.bot.MustOpen(s.ts.URL + "/popover_test")
	closed, err := s.bot.ClosePopover(".popover-close")
	s.Nil(err)
	s.Equal(2, closed)
}

func (s *BotClickSuite) TestMustClick() {
	s.T().Parallel()
	s.bot.MustOpen(s.ts.URL + "/click_test")
	s.NotPanics(func() {
		s.bot.MustClick("#clickme")
	})
	s.True(s.bot.page.MustHas("#clicked"))
}

func (s *BotClickSuite) TestClickError() {
	s.T().Parallel()
	s.bot.MustOpen(s.ts.URL + "/click_test")
	err := s.bot.Click("#non-existent")
	s.ErrorIs(err, context.DeadlineExceeded)
}

func (s *BotClickSuite) TestMustClickElemAndWait() {
	s.T().Parallel()
	s.bot.MustOpen(s.ts.URL + "/click_test")
	elem := s.bot.page.MustElement("#delayed-button")
	s.bot.MustClickElemAndWait(elem)
	s.True(s.bot.page.MustHas("#delayed-result"))
}

func (s *BotClickSuite) TestMustClickElem() {
	s.T().Parallel()
	s.bot.MustOpen(s.ts.URL + "/click_test")
	elem := s.bot.page.MustElement("#clickme")
	s.NotPanics(func() {
		s.bot.MustClickElem(elem)
	})
	s.True(s.bot.page.MustHas("#clicked"))
}

func (s *BotClickSuite) TestMustPressEscape() {
	s.T().Parallel()
	s.bot.MustOpen(s.ts.URL + "/click_test")
	s.NotPanics(func() {
		s.bot.MustPressEscape(s.bot.page.MustElement("body"))
	})
}

func (s *BotClickSuite) TestClickWithScript() {
	s.T().Parallel()
	s.bot.MustOpen(s.ts.URL + "/click_test")
	err := s.bot.ClickWithScript("#js-only-button")
	s.Nil(err)
	s.True(s.bot.page.MustHas("#js-clicked"))
}

func (s *BotClickSuite) TestMustAcceptCookies() {
	s.T().Parallel()
	s.bot.MustOpen(s.ts.URL + "/cookie_test")
	s.NotPanics(func() {
		s.bot.MustAcceptCookies("#accept-cookies", "#accept-all")
	})
	s.True(s.bot.page.MustHas("#cookies-accepted"))
}

func (s *BotClickSuite) TestAcceptCookiesError() {
	s.T().Parallel()
	s.bot.MustOpen(s.ts.URL + "/cookie_test")
	err := s.bot.AcceptCookies("#non-existent-cookie-button")
	s.Error(err)
	s.ErrorIs(err, ErrNoSelectorClicked)
}

func (s *BotClickSuite) TestMustClickSequentially() {
	s.T().Parallel()
	s.bot.MustOpen(s.ts.URL + "/sequential_click_test")
	s.NotPanics(func() {
		s.bot.MustClickSequentially("#button1", "#button2", "#button3")
	})
	s.True(s.bot.page.MustHas("#all-clicked-in-order"))
}
