package mini

import (
	"strings"

	"wee"

	"github.com/rs/zerolog/log"
)

type HtmlExtractor struct {
	*Shadow
}

func NewHtmlExtractor(s *Shadow) *HtmlExtractor {
	return &HtmlExtractor{
		Shadow: s,
	}
}

func (c *HtmlExtractor) ResultsCount() (string, float64) {
	sel := c.Mapper.ResultsCount
	if sel == "" {
		return "", 0
	}

	rc := c.Kwargs.ResultsCount
	txt := c.Bot.MustElemAttr(sel, wee.WithAttr(rc.Attr), wee.WithIndex(rc.Index))

	if rc.AttrSep != "" {
		txt = strings.Split(txt, rc.AttrSep)[rc.AttrIndex]
	}

	return txt, wee.MustStrToFloat(txt, rc.CharsAllowed)
}

func (c *HtmlExtractor) ResultsLoaded() int {
	count := 0

	for _, s := range c.Mapper.HasResults {
		elems, err := c.Bot.Elems(s)
		if err != nil {
			log.Error().Str("selector", s).Msg("cannot get elems")
			continue
		}

		count += len(elems)
	}

	return count
}
