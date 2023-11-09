package wee

import (
	"errors"
	"fmt"
)

var (
	ErrEmptySelector  = errors.New("empty selector")
	ErrGetElemsByText = errors.New("get elems not support by text")
	CannotFindElem    = errors.New("cannot find elem of selector")
)

func ErrCannotFindElem(sel string) error {
	return fmt.Errorf("Cannot find elem on selector %w : %s", CannotFindElem, sel)
}
