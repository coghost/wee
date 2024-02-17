package mini

type (
	Requester interface {
		MockInput(arr ...string) SerpStatus
		GotoNextPage() error
		MockHumanWait()
	}

	Responser interface {
		PageType() pageType
		PageData() *string
	}

	Extractor interface {
		ResultsCount() (string, float64)
		ResultsLoaded() int
	}
)

type pageType string

const (
	_pageTypeHTML pageType = "html"
	_pageTypeJSON pageType = "json"
)
