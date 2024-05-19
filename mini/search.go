package mini

import "errors"

type SearchStatus struct {
	URL     string
	PageNum int
}

var errNoSearchManger = errors.New("no search manager found.")

func (c *Crawler) isSearchFinished() bool {
	st, _ := c.getSearchStatus()
	if st.PageNum == c.scheme.Kwargs.MaxPages {
		return true
	}

	return false
}

func (c *Crawler) getSearchStatus() (*SearchStatus, error) {
	if c.SearchManager == nil {
		return nil, errNoSearchManger
	}

	return c.SearchManager.GetStatus(c.searchID)
}

func (c *Crawler) setSearchStatus(st *SearchStatus) error {
	if c.SearchManager == nil {
		return errNoSearchManger
	}

	// after directive is consumed, update status and reset start time.
	return c.SearchManager.SetStatus(c.searchID, st)
}
