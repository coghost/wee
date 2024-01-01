package mini

import (
	"wee"
	"wee/schemer"
)

type Shadow struct {
	Bot    *wee.Bot
	Scheme *schemer.Scheme
	Mapper *schemer.Mapper
	Kwargs *schemer.Kwargs
}

func NewShadow(bot *wee.Bot, scheme *schemer.Scheme) *Shadow {
	return &Shadow{
		Bot:    bot,
		Scheme: scheme,
		Mapper: scheme.Mapper,
		Kwargs: scheme.Kwargs,
	}
}
