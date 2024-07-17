package wee

import (
	"errors"
	"fmt"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/coghost/wee/fixtures"
	"github.com/go-rod/rod"
	"github.com/stretchr/testify/suite"
	"github.com/ungerik/go-dry"
)

type BotSuite struct {
	suite.Suite
	ts *httptest.Server
}

func TestBot(t *testing.T) {
	suite.Run(t, new(BotSuite))
}

func (s *BotSuite) SetupSuite() {
	s.ts = fixtures.NewTestServer()
}

func (s *BotSuite) TearDownSuite() {
}

func (s *BotSuite) Test00NewBot() {
	s.T().Parallel()

	want := "<html><head></head>\n<body>\nhellowee\n\n\n</body></html>"
	l, brw := NewBrowser()

	tests := []struct {
		name  string
		c     *Bot
		wantS string
	}{
		{
			name:  "new_bot",
			c:     NewBot(Headless(true), Launcher(l), Browser(brw)),
			wantS: want,
		},
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

			tt.c.MustOpen(s.ts.URL)
			raw := tt.c.page.MustHTML()
			raw = strings.ReplaceAll(raw, " ", "")

			s.Equal(tt.wantS, raw, "")
		})
	}
}

func (s *BotSuite) Test00BotWithOptionsOnly() {
	s.T().Parallel()

	bot := NewBotWithOptionsOnly()
	defer bot.Cleanup()

	s.Nil(bot.page)
	s.Panics(func() {
		bot.MustOpen(s.ts.URL)
	}, "new bot only with options panics when open url")

	BindBotLanucher(bot)

	s.NotNil(bot.page)
	s.NotPanics(func() {
		bot.MustOpen(s.ts.URL)
	}, "after binding bot with browser, shouldn't panic")
}

func (s *BotSuite) Test01CustomPageWithUAAndLang() {
	s.T().Parallel()

	userAgent := `wee - https://github.com/coghost/wee`
	lang := `en-CN,en`
	headers := []string{}

	l, brw := NewBrowser(
		BrowserNoDefaultDevice(false),
	)

	bot := NewBot(
		UserAgent(userAgent),
		AcceptLanguage(lang),
		Launcher(l),
		Browser(brw),
	)

	bot.MustOpen(s.ts.URL + "/headers")
	// bot.MustOpen("https://ip.42.pl/")

	elems := bot.page.MustElements(`ul>li`)
	for _, elem := range elems {
		headers = append(headers, elem.MustText())
	}

	s.Contains(headers, fmt.Sprintf("User-Agent: %s", userAgent), "user-agent must match")
	s.Contains(headers, fmt.Sprintf("Accept-Language: %s;q=0.9", lang), "accept-language must match")
}

const (
	_testCkStr = `sessionid=sessionid001; tz=Asia/Shanghai; logged_in=yes`
)

func (s *BotSuite) Test02CopyAsCURLCookies() {
	s.T().Parallel()

	tests := []struct {
		name  string
		ckKey string
		want  string
	}{
		{
			name:  "lowercase key",
			ckKey: "cookie",
			want:  _testCkStr,
		},
		{
			name:  "Canonical key",
			ckKey: "Cookie",
			want:  _testCkStr,
		},
	}
	for _, tt := range tests {
		raw := fmt.Sprintf(`curl 'http://127.0.0.1' -H '%s: %s'`, tt.ckKey, _testCkStr)

		bot := NewBotDefault(
			Headless(true),
			CopyAsCURLCookies([]byte(raw)),
		)

		bot.MustOpen(s.ts.URL + "/check_cookie")
		got := bot.page.MustElement(`pre`).MustText()
		s.Equal(tt.want, got, tt.name)
	}
}

func (s *BotSuite) Test02DumpCookies() {
	s.T().Parallel()

	bot := NewBotDefault()
	bot.MustOpen(s.ts.URL + "/set_cookie")

	ckFile, err := bot.DumpCookies()
	s.Nil(err)

	defer func() {
		err := os.Remove(ckFile)
		s.Nil(err)
	}()

	nodes, err := bot.LoadCookies(ckFile)
	s.Nil(err)
	s.Len(nodes, 3)

	wanted := strings.Split(_testCkStr, "; ")
	got := flattenNodes(nodes)

	s.ElementsMatch(wanted, got)
}

// const _ip42pl = `https://ip.42.pl/`

func (s *BotSuite) Test02WithCookies() {
	s.T().Parallel()

	baseURL := "https://www.baidu.com"

	tests := []struct {
		name string
		bot  *Bot
		want string
	}{
		{
			name: "with cookies",
			bot:  NewBotDefault(WithCookies(true)),
			want: `.cookies/https_www.baidu.com.cookies`,
		},
		{
			name: "with cookie file",
			bot:  NewBotDefault(WithCookieFile("/tmp/specified_cookie")),
			want: "/tmp/specified_cookie",
		},
		{
			name: "with cookie folder",
			bot:  NewBotDefault(WithCookieFolder("/tmp/specified")),
			want: "/tmp/specified/https_www.baidu.com.cookies",
		},
	}

	for _, tt := range tests {
		s.T().Run("test_"+tt.name, func(t *testing.T) {
			t.Parallel()
			tt.bot.MustOpen(baseURL)
			ckFile, err := tt.bot.DumpCookies()
			s.Nil(err)

			defer func() {
				err := os.Remove(ckFile)
				s.Nil(err)
			}()

			s.Equal(tt.want, ckFile)
		})
	}
}

const _ip42pl = `https://ip.42.pl/`

func getProxy() []string {
	file := "/tmp/p1.html"
	body, err := dry.FileGetString(file)
	if err != nil {
		bot := NewBotDefault()
		defer bot.Cleanup()
		bot.MustOpen("https://www.proxynova.com/proxy-server-list/")

		body = bot.page.MustHTML()
		_ = dry.FileSetString(file, body)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	dry.PanicIfErr(err)

	got := []string{}
	doc.Find("#tbl_proxy_list tr[data-proxy-id]").Each(func(i int, s *goquery.Selection) {
		port := s.Find("tr[data-proxy-id]>td[title]+td").Text()
		ip := s.Find("tr[data-proxy-id]>td[title]").After(`script`).Text()

		ip = strings.TrimSpace(ip)
		port = strings.TrimSpace(port)

		got = append(got, ip+":"+port)
	})

	return got
}

func (s *BotSuite) Test03Get() {
	getProxy()
}

func (s *BotSuite) ProxyServer() {
	// free proxy:
	// https://www.proxynova.com/proxy-server-list/
	// https://geonode.com/free-proxy-list

	proxyArr := getProxy()
	if len(proxyArr) == 0 {
		panic("cannot get proxy")
	}

	for _, proxy := range proxyArr {
		fmt.Printf("testing with %s\n", proxy)

		l, brw := NewBrowser(BrowserProxy(proxy))
		bot := NewBot(Launcher(l), Browser(brw))

		err := bot.Open(_ip42pl)

		if errors.Is(err, &rod.NavigationError{Reason: "net::ERR_CONNECTION_RESET"}) {
			bot.Cleanup()

			fmt.Printf("cannot connect: %s\n", proxy)
			RandSleepNap()
			continue
		}

		txt := bot.MustElemAttr(`p.ip`)
		if strings.Split(proxy, ":")[0] == txt {
			break
		}
	}

	QuitOnTimeout(2)
}

func (s *BotSuite) AuthProxyServer() {
	proxy := "ip:port"

	l, brw := NewBrowser(BrowserProxy(proxy))
	bot := NewBot(Launcher(l), Browser(brw))
	go brw.MustHandleAuth("username", "password")()

	bot.MustOpen(_ip42pl)

	ipaddr := bot.MustElemAttr(`p.ip`)
	fmt.Printf("got ip: %s\n", ipaddr)

	QuitOnTimeout(2)
}
