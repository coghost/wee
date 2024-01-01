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

	"github.com/enriquebris/goconcurrentqueue"
)

type SerpStatus int

const (
	SerpFailed SerpStatus = iota
	SerpOk
	SerpNoJobs
)

const TimeoutPerPageSeconds = 100

type (
	Searcher interface {
		MockOpen() bool
		MockInput(arr ...string) SerpStatus
	}

	Responser interface {
		ResponsePageData() *string
		GotoNextPage() error
		MockHumanWait()
	}

	Extractor interface {
		ParseResultsCount() (string, float64)
	}
)

type Crawler struct {
	searcher  Searcher
	extractor Extractor
	responser Responser
	storage   storage.Backend
	Queue     IQueue

	site       int
	searchTerm string
	status     SerpStatus
	startTime  time.Time

	maxPages int
	pageNum  int

	resultsCount  float64
	resultsLoaded float64

	filepath string

	bot    *wee.Bot
	kwargs *schemer.Kwargs

	// startPageNum by default set to 1
	startPageNum int
}

func NewCrawler(site int, search Searcher, resp Responser, extract Extractor, store storage.Backend) *Crawler {
	return &Crawler{
		searcher:  search,
		responser: resp,
		extractor: extract,
		storage:   store,

		site:         site,
		startPageNum: 1,
	}
}

func NewDefaultCrawler(site int, bot *wee.Bot, scheme *schemer.Scheme) *Crawler {
	shadow := NewShadow(bot, scheme)
	sch := &ClickSearcher{Shadow: shadow}
	resp := &Response{Shadow: shadow}
	extractor := &HtmlExtractor{Shadow: shadow}
	backend := storage.NewLocalFilesystemBackend("/tmp")
	qname := goconcurrentqueue.NewFIFO()

	return &Crawler{
		site:     site,
		maxPages: sch.Kwargs.MaxPages,

		bot:    bot,
		kwargs: scheme.Kwargs,

		searcher:  sch,
		responser: resp,
		extractor: extractor,
		storage:   backend,

		Queue: qname,

		startPageNum: 1,
	}
}

func (c *Crawler) String() string {
	return fmt.Sprintf("[SITE-%d]: %s, want(%d), start@%d, end@%d", c.site, c.searchTerm, c.maxPages, c.startPageNum, c.pageNum)
}

func (c *Crawler) GetSerps(keywords ...string) {
	c.startTime = time.Now()

	c.searchTerm = strings.Join(keywords, "_")

	c.searcher.MockOpen()

	if c.searcher.MockInput(keywords...) != SerpOk {
		return
	}

	txt, count := c.extractor.ParseResultsCount()
	c.resultsCount = count
	log.Info().Str("count_txt", txt).Float64("count", count).Msg("got number of results")

	startPage := c.startPageNum
	maxPages := c.kwargs.MaxPages + startPage

	success := make(chan bool)

	for i := startPage; i < maxPages; i++ {
		c.pageNum = i

		go func() {
			success <- c.getSerp(maxPages)
		}()

		select {
		case res := <-success:
			if !res {
				break
			}
		case <-time.After(time.Duration(TimeoutPerPageSeconds) * time.Second):
			break
		}
	}
}

func (c *Crawler) getSerp(maxPages int) bool {
	log.Debug().Int("page_num", c.pageNum).Msg("crawling page")

	pageData := c.responser.ResponsePageData()
	c.SaveCaptured(pageData)

	if maxPages-1 == c.pageNum {
		// break when final page reached
		c.bot.MustScrollToBottom()
		return false
	}

	err := c.responser.GotoNextPage()
	if err != nil {
		return false
	}

	c.responser.MockHumanWait()
	// reset start time
	c.startTime = time.Now()

	return true
}

func (c *Crawler) uniqName() string {
	name := fmt.Sprintf("%s_%s_%d", getUniqid(), c.searchTerm, c.pageNum)
	name = strings.ReplaceAll(name, "/", "-")

	return name
}

func (c *Crawler) SaveCaptured(content *string) bool {
	if content == nil {
		return true
	}

	pageType := _pageTypeHTML
	if IsJSON(*content) {
		pageType = _pageTypeJSON
	}

	filePath := fmt.Sprintf("%d/%s.%s", c.site, c.uniqName(), pageType)
	c.filepath = filePath

	if err := c.storage.PutObject(filePath, []byte(*content)); err != nil {
		log.Error().Err(err).Str("filepath", filePath).Msg("cannot save captured data")
		return false
	}

	sd := c.newSerpDirective(content)

	raw, err := wee.Stringify(sd)
	if err != nil {
		log.Error().Err(err).Msg("cannot stringify serp directive")
		return false
	}

	fmt.Println(raw)
	c.Queue.Enqueue(raw)

	return true
}

// newSerpDirective.
func (c *Crawler) newSerpDirective(raw *string) *SerpDirective {
	st := strings.ReplaceAll(c.searchTerm, ">", "_")
	serpId := fmt.Sprintf("%d_%s_%s", c.site, xdtm.UTCNow().ToShortDateTimeString(), st)

	return &SerpDirective{
		Site:     c.site,
		PageNum:  c.pageNum,
		PageSize: len(*raw),

		ResultsLoaded: c.resultsLoaded,
		ResultsCount:  c.resultsCount,

		TimeOnPage: time.Since(c.startTime).Seconds(),
		Filepath:   c.filepath,
		SerpUrl:    c.bot.CurrentUrl(),
		SearchTerm: c.searchTerm,
		SerpID:     serpId,
	}
}

type SerpDirective struct {
	Site     int `json:"site"`
	PageNum  int `json:"page_num"`
	PageSize int `json:"page_size"`

	ResultsLoaded float64 `json:"results_loaded"`
	ResultsCount  float64 `json:"results_count"`

	TimeOnPage float64 `json:"time_on_page"`

	Filepath   string `json:"filepath"`
	SerpUrl    string `json:"serp_url"`
	SearchTerm string `json:"search_term"`
	SerpID     string `json:"serp_id"`
}
