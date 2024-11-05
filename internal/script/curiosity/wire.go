//go:build wireinject
// +build wireinject

package curiosity

import (
	"github.com/google/wire"

	"github.com/danielmesquitta/openai-web-scraper/internal/config"
	"github.com/danielmesquitta/openai-web-scraper/internal/pkg/validator"
	"github.com/danielmesquitta/openai-web-scraper/internal/provider/scraper"
	"github.com/danielmesquitta/openai-web-scraper/internal/provider/scraper/pw"
)

func NewGenerateCuriosities() *GenerateCuriosities {
	wire.Build(
		wire.Bind(new(validator.Validator), new(*validator.Validate)),
		validator.NewValidate,

		config.LoadEnv,

		wire.Bind(new(scraper.Scraper), new(*pw.PlayWrightScraper)),
		pw.NewPlayWrightScraper,

		newGenerateCuriosities,
	)

	return &GenerateCuriosities{}
}
