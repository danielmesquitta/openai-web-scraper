package pw

import (
	"github.com/danielmesquitta/openai-web-scraper/internal/config"
	"github.com/danielmesquitta/openai-web-scraper/internal/provider/scraper"
)

type PlayWrightScraper struct {
	e *config.Env
}

func NewPlayWrightScraper(
	e *config.Env,
) *PlayWrightScraper {
	return &PlayWrightScraper{
		e: e,
	}
}

var _ scraper.Scraper = &PlayWrightScraper{}
