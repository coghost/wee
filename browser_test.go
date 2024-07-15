package wee

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/go-rod/rod/lib/proto"
	"github.com/samber/lo"
	"github.com/stretchr/testify/suite"
	"github.com/ungerik/go-dry"
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

func (s *LaunchSuite) TestParseCookie() {
	raw, err := dry.FileGetString("/tmp/cURL.txt")
	s.Nil(err)
	cks, err := ParseCopyAsCURLAsCookie(raw)
	s.Nil(err)

	v, err := json.Marshal(cks)
	s.Nil(err)

	err = dry.FileSetBytes("/tmp/curl.cookies", v)
	s.Nil(err)
}

func (s *LaunchSuite) Test_01_cookie() {
	raw := `views_132689=yes; wordpress_test_cookie=WP+Cookie+check; wordpress_logged_in_8f24464e4a1e05c313cf60c04d13d274=Hex%7C1700997303%7CIKXetb9KHNR7bN15AebHi0kXgwYtfX5NOEzCwEIF9gZ%7Cea11c7f4a05a7b0750776ce1a29f7c47c1330e46c61a412df381c6fd45db18e9; wp-settings-time-1448=1700824503; views_131771=yes; views_129606=yes; views_129359=yes; views_124189=yes; views_124432=yes; views_117539=yes; views_13794=yes; PHPSESSID=atv5rpufpivo360oppdm4vhggu`
	domain := `example.com`

	var ckarr []map[string]interface{}

	arr := lo.Map(strings.Split(raw, ";"),
		func(item string, index int) string {
			return strings.TrimSpace(item)
		})

	for _, line := range arr {
		d := make(map[string]interface{})
		ar := strings.Split(line, "=")
		if strings.Contains(ar[0], "views_") {
			continue
		}
		d["name"] = ar[0]
		d["value"] = ar[1]
		d["domain"] = domain
		d["path"] = "/"
		d["source_port"] = 443

		ckarr = append(ckarr, d)
	}

	fmt.Println(ckarr)
	err := dry.FileSetJSON("/tmp/001.json", ckarr)
	s.Nil(err)

	// val, _ := dry.FileGetBytes("/tmp/backup_https_nfcloud.net.cookies")
	val, _ := dry.FileGetBytes("/tmp/001.txt")
	var cookies []proto.NetworkCookie
	err = json.Unmarshal(val, &cookies)
	s.Nil(err)
	fmt.Println(cookies)
}
