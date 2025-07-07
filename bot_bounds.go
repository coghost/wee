package wee

func WindowMaximize(b bool) BotOption {
	return func(o *Bot) {
		o.windowMaximize = b
	}
}

func WithBounds(bounds *BrowserBounds) BotOption {
	return func(o *Bot) {
		o.bounds = bounds
	}
}

// BrowserBounds (experimental) Browser window bounds information.
type BrowserBounds struct {
	// Left (optional) The offset from the left edge of the screen to the window in pixels.
	Left int `json:"left,omitempty"`

	// Top (optional) The offset from the top edge of the screen to the window in pixels.
	Top int `json:"top,omitempty"`

	// Width (optional) The window width in pixels.
	Width int `json:"width,omitempty"`

	// Height (optional) The window height in pixels.
	Height int `json:"height,omitempty"`
}

func NewBounds(w, h int) *BrowserBounds {
	return &BrowserBounds{
		Left:   0,
		Top:    0,
		Width:  w,
		Height: h,
	}
}

func NewBoundsVGA480() *BrowserBounds {
	return NewBounds(640, 480) //nolint:mnd
}

// NewBoundsSVGA600
//
//	@return *BrowserBounds {0,0,800,640}
func NewBoundsSVGA600() *BrowserBounds {
	return NewBounds(800, 600) //nolint:mnd
}

func NewBoundsXVGA() *BrowserBounds {
	return NewBounds(1024, 768) //nolint:mnd
}

func NewBoundsSXGA() *BrowserBounds {
	return NewBounds(1280, 1024) //nolint:mnd
}

func NewBoundsHD() *BrowserBounds {
	return NewBounds(1280, 720) //nolint:mnd
}

func NewBoundsFullHD() *BrowserBounds {
	return NewBounds(1920, 1080) //nolint:mnd
}

func NewBounds2K() *BrowserBounds {
	return NewBounds(2560, 1440) //nolint:mnd
}

func NewBounds4K() *BrowserBounds {
	return NewBounds(3480, 2160) //nolint:mnd
}
