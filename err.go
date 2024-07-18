package wee

import (
	"errors"
	"fmt"
)

var (
	ErrEmptySelector  = errors.New("empty selector")
	ErrGetElemsByText = errors.New("get elems not support by text")
	ErrCannotFindElem = errors.New("cannot find elem of selector")
)

func ErrCannotFindSelector(sel string) error {
	return fmt.Errorf("cannot find elem on selector %w : %s", ErrCannotFindElem, sel)
}
