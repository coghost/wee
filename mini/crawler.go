package mini

import (
	"fmt"
	"strings"
	"time"

	"wee"
	"wee/schemer"

	"github.com/chartmuseum/storage"
	"github.com/coghost/xdtm"
	"github.com/rs/zerolog/log"
)

type SerpStatus int

const (
	SerpFailed SerpStatus = iota
	SerpOk
	SerpNoJobs
	SerpCannotOpenPage
	SerpCannotInput
)

const _perPageMaxDwellSeconds = 100

type Crawler struct {
	requester Requester
	extractor Extractor
	responser Responser
	// fileStorage where we store captured page.
	fileStorage storage.Backend

	site       int
	searchTerm string

	// searchID searchID is `<site>_<searchTerm>`
	searchID string

	startTime time.Time

	startPageNum  int // startPageNum by default set to 1
	remainedPages int

	pageNum      int
	resultsCount float64

	bot    *wee.Bot
	scheme *schemer.Scheme

	// SerpDirectiveChan is serp page directive info.
	SerpDirectiveChan chan *SerpDirective

	CapturedIdFunc func() string
}

func NewCrawler(site int, bot *wee.Bot, scheme *schemer.Scheme, requester Requester, resp Responser, extract Extractor, fileStore storage.Backend) *Crawler {
	crawl := &Crawler{
		site:   site,
		bot:    bot,
		scheme: scheme,

		requester: requester,
		responser: resp,
		extractor: extract,

		fileStorage: fileStore,

		// CapturedIdFunc
		CapturedIdFunc: func() string {
			return xdtm.UTCNow().ToShortDateTimeString()
		},
	}

	crawl.initialize()

	return crawl
}

func NewTextInputCrawler(site int, bot *wee.Bot, scheme *schemer.Scheme, store storage.Backend) *Crawler {
	shadow := NewShadow(scheme)
	req := NewRequest(shadow, bot, newTextInput(bot))
	resp := NewResponse(shadow, bot)
	extractor := NewHTMLExtractor(shadow, bot)

	return NewCrawler(site, bot, scheme, req, resp, extractor, store)
}

func (c *Crawler) Cleanup() {
	c.bot.Cleanup()
}

func (c *Crawler) initialize() {
	c.setPageNum(1)

	c.startTime = time.Now()
	c.SerpDirectiveChan = make(chan *SerpDirective)
}

func (c *Crawler) setPageNum(pageNums ...int) {
	n := wee.FirstOrDefault(1, pageNums...)
	c.startPageNum = n
}

func (c *Crawler) String() string {
	return fmt.Sprintf("[SITE-%d]: %s, want(%d), start@%d, end@%d", c.site, c.searchTerm, c.scheme.Kwargs.MaxPages, c.startPageNum, c.pageNum)
}

func (c *Crawler) ReloadPage(url string, pageNum int) error {
	err := c.requester.MockOpen(url)
	if err != nil {
		return err
	}

	c.setPageNum(pageNum)
	log.Info().Str("crawler", c.String()).Msgf("restarted")

	return nil
}

func (c *Crawler) bindSearch(keywords []string) {
	c.searchTerm = strings.Join(keywords, "_")
	c.searchID = fmt.Sprintf("%d:%s", c.site, c.searchTerm)
}

// GetSerps
//
//	@param interceptInput if has error, will do input, else will bypass input.
//	@param interceptAfterInput something after input.
func (c *Crawler) GetSerps(
	keywords []string,
	interceptInput func(string) error,
	interceptAfterInput func(string) error,
) {
	c.bindSearch(keywords)

	// TODO: move this out of crawler.
	wee.UpdateBotWithOption(c.bot)

	log.Info().Str("crawler", c.String()).Msgf("started")
	if err := c.requester.MockOpen(); err != nil {
		log.Fatal().Err(err).Msg("cannot open page.")
	}

	if err := interceptInput(c.searchID); err != nil {
		if c.requester.MockInput(keywords...) != SerpOk {
			log.Warn().Msg("cannot do mock input")
			return
		}
	}

	err := interceptAfterInput(c.searchID)
	if err != nil {
		log.Warn().Err(err).Msg("intercept failed")
	}

	var txt string
	txt, c.resultsCount = c.extractor.ResultsCount()
	log.Info().Str("count_txt", txt).Float64("count", c.resultsCount).Msg("got number of results")

	maxPages := c.scheme.Kwargs.MaxPages + 1

	for i := c.startPageNum; i < maxPages; i++ {
		c.pageNum = i
		log.Debug().Int("page_num", c.pageNum).Msg("crawling page")

		c.saveCaptured()

		if maxPages-1 == c.pageNum {
			// when max page reached, no need to load next page content,
			// we just scroll to bottom in case page contents are not fully loaded.
			c.bot.MustScrollToBottom(wee.WithHumanized(c.scheme.Ctrl.Humanized))
			break
		}

		if !c.gotoNextPageInTime(c.pageNum) {
			break
		}
	}

	c.SerpDirectiveChan <- nil
}

func NoIntercept(string) error {
	return nil
}

// ConsumeSerp
//
//	@param timeout the max time serp consumer can run
//	@param isTotalTimeout is `timeout` for total serp or just for one serp
func (c *Crawler) ConsumeSerp(consumer func(*SerpDirective), timeout time.Duration, isTotalTimeout bool) {
	pollTimeout := time.After(timeout)

	for {
		breakLoop := false

		select {
		case sd := <-c.SerpDirectiveChan:
			if sd == nil {
				log.Info().Msg("final page reached.")
				breakLoop = true

				break
			}

			consumer(sd)

			// reset the timer
			if !isTotalTimeout {
				pollTimeout = time.After(timeout)
			}
		case <-pollTimeout:
			log.Warn().Msg("max timeout reached.")
			breakLoop = true

			break
		}

		if breakLoop {
			break
		}
	}
}

func (c *Crawler) gotoNextPageInTime(pageNum int) bool {
	success := make(chan error)

	go func() {
		success <- c.gotoNextPage()
	}()

	select {
	case err := <-success:
		if err != nil {
			log.Error().Err(err).Int("page_num", pageNum).Msg("get serp failed")
			return false
		}
	case <-time.After(time.Duration(_perPageMaxDwellSeconds) * time.Second):
		log.Error().Int("page_num", pageNum).Msg("get serp timeout")
		return false
	}

	return true
}

func (c *Crawler) gotoNextPage() error {
	err := c.requester.GotoNextPage()
	if err != nil {
		return err
	}

	if c.scheme.Ctrl.Humanized {
		c.requester.MockHumanWait()
	}

	return nil
}

func (c *Crawler) saveCaptured() bool {
	content := c.responser.PageData()

	if content == nil {
		return true
	}

	pageType := c.responser.PageType()

	filePath := fmt.Sprintf("%d/%s.%s", c.site, c.uniqName(), pageType)

	if err := c.fileStorage.PutObject(filePath, []byte(*content)); err != nil {
		log.Error().Err(err).Str("filepath", filePath).Msg("cannot save captured data")
		return false
	}

	sd := c.newSerpDirective(filePath, content)
	// reset start time
	c.startTime = time.Now()

	c.SerpDirectiveChan <- sd

	// raw, err := wee.Stringify(sd)
	// if err != nil {
	// 	log.Error().Err(err).Msg("cannot stringify serp directive")
	// 	return false
	// }
	// c.Queue.Enqueue(raw)

	return true
}

func (c *Crawler) uniqName() string {
	// uid := fmt.Sprintf("%s_%f", xdtm.UTCNow().ToRfc3339MicroString(), rand.Float64())
	name := fmt.Sprintf("%s_%s_%d", c.CapturedIdFunc(), c.searchTerm, c.pageNum)
	name = strings.ReplaceAll(name, "/", "-")

	return name
}

func (c *Crawler) newSerpDirective(filepath string, raw *string) *SerpDirective {
	st := strings.ReplaceAll(c.searchTerm, ">", "_")
	serpId := fmt.Sprintf("%d_%s_%s", c.site, c.CapturedIdFunc(), st)

	return &SerpDirective{
		Site:     c.site,
		PageNum:  c.pageNum,
		PageSize: len(*raw),

		ResultsLoaded: c.extractor.ResultsLoaded(),
		ResultsCount:  c.resultsCount,

		PageDwellTime: time.Since(c.startTime).Seconds(),
		Filepath:      filepath,
		SerpUrl:       c.bot.CurrentUrl(),
		SearchTerm:    c.searchTerm,
		SerpID:        serpId,
		SearchID:      c.searchID,
	}
}

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
