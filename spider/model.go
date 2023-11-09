package spider

type Spider struct {
	Id   int32
	Uid  string
	Home string

	Selectors *Selectors
}

type Selectors struct {
	Popovers []string

	SortBy *Dropdown

	HasResults   []string
	NoResults    []string
	ResultsCount string

	Paginate string
	PageNum  string
}

type Dropdown struct {
	Trigger string
	Item    string
}

func NewDropdown(t, s string) *Dropdown {
	return &Dropdown{t, s}
}
