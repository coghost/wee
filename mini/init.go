package mini

import (
	"fmt"
	"time"

	"wee"
	"wee/schemer"

	"github.com/chartmuseum/storage"
	"github.com/coghost/xdtm"
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

const _startPageNum = 1

type Crawler struct {
	// FileStorage where we store captured page.
	FileStorage storage.Backend

	Inputer   Inputer
	Paginator Paginator

	SearchManager SearchManager
	// SerpIdNowFunc the same searchTerm maybe queried multiple times,
	// so to distinguish each result, we should tag it with the timestamp.
	SerpIdNowFunc func() string

	Consumer func(*SerpDirective)

	// bot
	bot    *wee.Bot
	scheme *schemer.Scheme

	// shadow *Shadow
	// requester *Request
	// responser *Response
	// extractor is how we extract all basic info of serp.
	// crawler is built upon go-rod, so whole HTML is loaded, we can use HTMLExtractor by default.
	// extractor *HTMLExtractor

	// Serp Info
	// site a unique id where we can tell one site from another.
	site int

	// searchTerm is unique for each site.
	searchTerm string
	// searchID is `<site>:<searchTerm>` globally unique.
	searchID string

	// startTime
	startTime time.Time
	// startPageNum by default set to 1
	startPageNum int
	// pageNum
	pageNum int

	// serpResultsRaw
	serpResultsRaw string
	// serpResultsCount
	serpResultsCount float64
	// pageResultsLoaded how many results are there in current page.
	pageResultsLoaded int

	// serpDirectiveChan is serp page directive info.
	serpDirectiveChan chan *SerpDirective
}

type CrawlerOption func(*Crawler)

func bindCrawlerOptions(c *Crawler, opts ...CrawlerOption) {
	for _, f := range opts {
		f(c)
	}
}

func WithSite(i int) CrawlerOption {
	return func(c *Crawler) {
		c.site = i
	}
}

func WithScheme(s *schemer.Scheme) CrawlerOption {
	return func(c *Crawler) {
		c.scheme = s
	}
}

func WithBot(b *wee.Bot) CrawlerOption {
	return func(c *Crawler) {
		c.bot = b
	}
}

// func WithInputer(in Inputer) CrawlerOption {
// 	return func(c *Crawler) {
// 		c.inputer = in
// 	}
// }

func WithFileStorage(s storage.Backend) CrawlerOption {
	return func(c *Crawler) {
		c.FileStorage = s
	}
}

func NewCrawler(site int, scheme *schemer.Scheme, options ...CrawlerOption) *Crawler {
	tmpFolder := storage.NewLocalFilesystemBackend("/tmp/minicc")
	crawler := &Crawler{FileStorage: tmpFolder}

	bindCrawlerOptions(crawler, options...)

	crawler.initialize()

	return crawler
}

func NewCrawlerInput(site int, scheme *schemer.Scheme, opts ...CrawlerOption) *Crawler {
	return NewCrawler(site, scheme, opts...)
}

func NewTextInputCrawler(site int, scheme *schemer.Scheme, bot *wee.Bot, store storage.Backend) *Crawler {
	crawler := &Crawler{
		site:   site,
		scheme: scheme,
		bot:    bot,

		FileStorage: store,
	}

	crawler.Inputer = newTextInput(bot)
	// crawler.Paginator =

	crawler.initialize()

	return crawler
}

func (c *Crawler) Cleanup() {
	c.initialize()
}

func (c *Crawler) String() string {
	return fmt.Sprintf("[%d]:%q, start/total/current/remained@%d/%d:%d/%d",
		c.site, c.searchTerm,
		c.startPageNum, c.scheme.Kwargs.MaxPages,
		c.pageNum, c.scheme.Kwargs.MaxPages-c.pageNum)
}

func (c *Crawler) initialize() {
	c.setPageNum()

	c.startTime = time.Now()
	c.serpDirectiveChan = make(chan *SerpDirective)
	c.SerpIdNowFunc = nowFunc
}

func (c *Crawler) setPageNum(pageNums ...int) {
	n := wee.FirstOrDefault(_startPageNum, pageNums...)
	c.startPageNum = n
}

var nowFunc = func() string {
	return xdtm.UTCNow().ToShortDateTimeString()
}

func PrintConsumer(sd *SerpDirective) {
	fmt.Printf("%+v\n", sd) //nolint
}
