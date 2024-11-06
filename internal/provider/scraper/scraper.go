package scraper

import "context"

type OpenAIModel string

const (
	OpenAIModelO1Preview OpenAIModel = "o1-preview"
	OpenAIModelO1Mini    OpenAIModel = "o1-mini"
	OpenAIModel4oMini    OpenAIModel = "gpt-4o-mini"
	OpenAIModel4o        OpenAIModel = "gpt-4o"
)

type Scraper interface {
	ScrapOpenAIPrompt(
		ctx context.Context,
		prompt string,
		model OpenAIModel,
	) (string, error)
}
