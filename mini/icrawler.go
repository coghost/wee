package mini

type (
	Inputer interface {
		Input(selectors []string, arr ...string) error
	}

	Paginator interface {
		GotoNextPage() error
	}

	SearchManager interface {
		GetStatus(string) (*SearchStatus, error)
		SetStatus(string, *SearchStatus) error
	}

	// 	Requester interface {
	// 		MockOpen(urls ...string) error
	// 		MockInput(arr ...string) SerpStatus
	// 		GotoNextPage() error
	// 		MockHumanWait()
	// 	}
	// Responser interface {
	// 	PageType() PageType
	// 	PageData() (PageType, *string)
	// }
	// Extractor interface {
	// 	ResultsCount() (string, float64)
	// 	ResultsLoaded() int
	// }
)

type PageType string

const (
	PageTypeHTML PageType = "html"
	PageTypeJSON PageType = "json"
)

type SerpDirective struct {
	Site     int `json:"site,omitempty"`
	PageNum  int `json:"page_num,omitempty"`
	PageSize int `json:"page_size,omitempty"`

	ResultsLoaded int     `json:"results_loaded,omitempty"`
	ResultsCount  float64 `json:"results_count,omitempty"`

	PageDwellTime float64 `json:"page_dwell_time,omitempty"`

	Filepath   string `json:"filepath,omitempty"`
	SerpUrl    string `json:"serp_url,omitempty"`
	SearchTerm string `json:"search_term,omitempty"`
	SerpID     string `json:"serp_id,omitempty"`
	SearchID   string `json:"search_id,omitempty"`
}
