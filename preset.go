package wee

const (
	SEP       = "@@@"
	IFrameSep = "$$$"
)

const (
	// ZeroToSec no timeout in second
	ZeroToSec = 0
	// NapToSec a nap timeout in second
	NapToSec = 2
	// ShortToSec short timeout in second
	ShortToSec = 10
	// MediumToSec medium timeout in second
	MediumToSec = 20
	// LongToSec long timeout in second
	LongToSec = 60
)

const (
	PT1Sec  = 1
	PT10Sec = 10
	PT20Sec = 20
	PT60Sec = 60
	// PT10MilliSec a very short timeout in millisecond
	PT10MilliSec = 0.01
)

type PanicByType int

const (
	PanicByDft PanicByType = iota
	PanicByDump
	PanicByLogError
)

const (
	LongScrollStep     = 32
	MediumScrollStep   = 16
	ShortScrollStep    = 8
	QuickScrollStep    = 4
	DirectlyScrollStep = 1
)
