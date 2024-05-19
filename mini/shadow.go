package mini

import (
	"wee/schemer"
)

type Shadow struct {
	Scheme *schemer.Scheme
	Mapper *schemer.Mapper
	Kwargs *schemer.Kwargs
}

func NewShadow(scheme *schemer.Scheme) *Shadow {
	return &Shadow{
		Scheme: scheme,
		Mapper: scheme.Mapper,
		Kwargs: scheme.Kwargs,
	}
}
