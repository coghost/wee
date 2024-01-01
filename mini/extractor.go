package mini

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	"wee"

	"github.com/coghost/xdtm"
	"github.com/spf13/cast"
)

const (
	_pageTypeHTML = "html"
	_pageTypeJSON = "json"
)

type HtmlExtractor struct {
	*Shadow
}

func (c *HtmlExtractor) ParseResultsCount() (string, float64) {
	sel := c.Mapper.ResultsCount
	if sel == "" {
		return "", 0
	}

	rc := c.Kwargs.ResultsCount
	txt := c.Bot.MustElemAttr(sel, wee.WithAttr(rc.Attr), wee.WithIndex(rc.Index))

	if rc.AttrSep != "" {
		txt = strings.Split(txt, rc.AttrSep)[rc.AttrIndex]
	}

	return txt, MustStrToFloat(txt, rc.CharsAllowed)
}

func StrToNumChars(raw string, keptChars string) string {
	chars := "[0-9" + keptChars + "]+"
	re := regexp.MustCompile(chars)
	c := re.FindAllString(raw, -1)
	r := strings.Join(c, "")

	return r
}

func MustStrToFloat(raw string, keptChars string) float64 {
	v := StrToNumChars(raw, keptChars)
	return cast.ToFloat64(v)
}

func IsJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

func getUniqid() string {
	name := fmt.Sprintf("%s_%f", xdtm.UTCNow().ToRfc3339MicroString(), rand.Float64())
	return name
}
