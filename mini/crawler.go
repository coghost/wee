package mini

import (
	"fmt"
	"math/rand"
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

const PerPageMaxDwellSeconds = 100

type Crawler struct {
	requester Requester
	extractor Extractor
	responser Responser

	storage storage.Backend
	Queue   IQueue

	site         int
	searchTerm   string
	startTime    time.Time
	startPageNum int // startPageNum by default set to 1
	pageNum      int
	resultsCount float64

	bot    *wee.Bot
	scheme *schemer.Scheme
}

func NewCrawler(site int, bot *wee.Bot, scheme *schemer.Scheme, requester Requester, resp Responser, extract Extractor, store storage.Backend, queue IQueue) *Crawler {
	return &Crawler{
		site: site,

		bot:    bot,
		scheme: scheme,

		requester: requester,
		responser: resp,
		extractor: extract,
		storage:   store,
		Queue:     queue,

		startPageNum: 1,
		startTime:    time.Now(),
	}
}

func NewTextInputCrawler(site int, bot *wee.Bot, scheme *schemer.Scheme) *Crawler {
	shadow := NewShadow(bot, scheme)
	req := NewRequest(shadow, newTextInput(bot))
	resp := NewResponse(shadow)
	extractor := NewHtmlExtractor(shadow)
	backend := storage.NewLocalFilesystemBackend("/tmp")
	fifo := goconcurrentqueue.NewFIFO()

	return NewCrawler(site, bot, scheme, req, resp, extractor, backend, fifo)
}

func (c *Crawler) String() string {
	return fmt.Sprintf("[SITE-%d]: %s, want(%d), start@%d, end@%d", c.site, c.searchTerm, c.scheme.Kwargs.MaxPages, c.startPageNum, c.pageNum)
}

func (c *Crawler) GetSerps(keywords ...string) {
	c.searchTerm = strings.Join(keywords, "_")

	if c.requester.MockInput(keywords...) != SerpOk {
		return
	}

	var txt string
	txt, c.resultsCount = c.extractor.ResultsCount()
	log.Info().Str("count_txt", txt).Float64("count", c.resultsCount).Msg("got number of results")

	maxPages := c.scheme.Kwargs.MaxPages + c.startPageNum

	for i := c.startPageNum; i < maxPages; i++ {
		c.pageNum = i
		log.Debug().Int("page_num", c.pageNum).Msg("crawling page")

		c.saveCaptured()

		if maxPages-1 == c.pageNum {
			// when max page reached, no need to load next page content,
			// we just scroll to bottom in case page contents are not fully loaded.
			c.bot.MustScrollToBottom(wee.WithHumanized(true))
			break
		}

		if !c.gotoNextPageInTime(c.pageNum) {
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
	case <-time.After(time.Duration(PerPageMaxDwellSeconds) * time.Second):
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

	c.requester.MockHumanWait()

	return nil
}

func (c *Crawler) saveCaptured() bool {
	content := c.responser.PageData()

	if content == nil {
		return true
	}

	pageType := c.responser.PageType()

	filePath := fmt.Sprintf("%d/%s.%s", c.site, c.uniqName(), pageType)

	if err := c.storage.PutObject(filePath, []byte(*content)); err != nil {
		log.Error().Err(err).Str("filepath", filePath).Msg("cannot save captured data")
		return false
	}

	sd := c.newSerpDirective(filePath, content)
	// reset start time
	c.startTime = time.Now()

	raw, err := wee.Stringify(sd)
	if err != nil {
		log.Error().Err(err).Msg("cannot stringify serp directive")
		return false
	}

	c.Queue.Enqueue(raw)

	return true
}

func (c *Crawler) uniqName() string {
	uid := fmt.Sprintf("%s_%f", xdtm.UTCNow().ToRfc3339MicroString(), rand.Float64())
	name := fmt.Sprintf("%s_%s_%d", uid, c.searchTerm, c.pageNum)
	name = strings.ReplaceAll(name, "/", "-")

	return name
}

// newSerpDirective.
func (c *Crawler) newSerpDirective(filepath string, raw *string) *SerpDirective {
	st := strings.ReplaceAll(c.searchTerm, ">", "_")
	serpId := fmt.Sprintf("%d_%s_%s", c.site, xdtm.UTCNow().ToShortDateTimeString(), st)

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
	}
}

type SerpDirective struct {
	Site     int `json:"site"`
	PageNum  int `json:"page_num"`
	PageSize int `json:"page_size"`

	ResultsLoaded int     `json:"results_loaded"`
	ResultsCount  float64 `json:"results_count"`

	PageDwellTime float64 `json:"page_dwell_time"`

	Filepath   string `json:"filepath"`
	SerpUrl    string `json:"serp_url"`
	SearchTerm string `json:"search_term"`
	SerpID     string `json:"serp_id"`
}
