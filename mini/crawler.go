package mini

import (
	"time"

	"github.com/rs/zerolog/log"
)

var perPage1Min = time.Second * 60 //nolint: gomnd

// GetSerps gets all pages of required.
func (c *Crawler) GetSerps(keywords ...string) {
	c.bindSearch(keywords)
	log.Logger = log.With().Int("site", c.site).Str("search_term", c.searchTerm).Logger()

	done := make(chan bool)

	if c.Consumer == nil {
		log.Info().Msg("use default print consumer")
		c.Consumer = PrintConsumer
	}

	go func() {
		c.ConsumeSerps(c.Consumer, perPage1Min)
		done <- true
	}()

	if c.isSearchFinished() {
		log.Debug().Msg("already finished")
		c.serpDirectiveChan <- nil

		return
	}

	c.getSerps(keywords)

	<-done
}

func (c *Crawler) getSerps(keywords []string) {
	if err := c.MockOpen(); err != nil {
		log.Fatal().Err(err).Msg("cannot open page.")
	}

	if c.MockInput(keywords...) != SerpOk {
		log.Warn().Msg("cannot do mock input")
		return
	}

	c.serpResultsRaw, c.serpResultsCount = c.parseResultsCount()
	log.Debug().Str("count_raw", c.serpResultsRaw).Float64("count", c.serpResultsCount).Msg("number of serp results")

	for i := c.startPageNum; i <= c.scheme.Kwargs.MaxPages; i++ {
		c.pageNum = i
		log.Debug().Msgf("crawler: %s", c.String())
		c.scrollDownToBottom()

		directive, err := c.saveCaptured()
		if err != nil {
			break
		}

		c.serpDirectiveChan <- directive
		c.setSearchStatus(&SearchStatus{
			URL:     directive.SerpUrl,
			PageNum: directive.PageNum,
		})

		// reset start time
		c.startTime = time.Now()

		if c.scheme.Kwargs.MaxPages == c.pageNum {
			// when max page reached, no need to load next page content
			break
		}

		if !c.gotoNextPageInTime(c.pageNum) {
			break
		}

		if c.scheme.Ctrl.Humanized {
			c.waitInPagination()
		}

		// wait go to next ready
		c.bot.MustWaitLoad()
	}

	c.serpDirectiveChan <- nil
}

// ConsumeSerps
//
//	@param timeout the max time of getting serp before quit.
func (c *Crawler) ConsumeSerps(consumer func(*SerpDirective), timeout time.Duration) {
	pollTimeout := time.After(timeout)

	for {
		breakLoop := false

		select {
		case sd := <-c.serpDirectiveChan:
			if sd == nil {
				log.Info().Msg("final page reached.")
				breakLoop = true

				break
			}

			consumer(sd)

			pollTimeout = time.After(timeout)
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
