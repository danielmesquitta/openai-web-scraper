package pw

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/playwright-community/playwright-go"

	"github.com/danielmesquitta/openai-web-scraper/internal/provider/scraper"
)

func (pws *PlayWrightScraper) ScrapOpenAIPrompt(
	ctx context.Context,
	prompt string,
	model scraper.OpenAIModel,
) (string, error) {
	select {
	case <-ctx.Done():
		return "", errors.New("context canceled or deadline exceeded")
	default:
		pw, err := playwright.Run()
		if err != nil {
			return "", fmt.Errorf("could not start playwright: %w", err)
		}
		defer func() { _ = pw.Stop() }()

		cdpURL := fmt.Sprintf("http://localhost:%s", pws.e.CDPPort)
		browser, err := pw.Chromium.ConnectOverCDP(cdpURL)
		if err != nil {
			return "", fmt.Errorf("could not launch browser: %w", err)
		}
		contexts := browser.Contexts()
		if len(contexts) == 0 {
			return "", fmt.Errorf("could not get contexts")
		}
		page, err := contexts[0].NewPage()
		if err != nil {
			return "", fmt.Errorf("could not create page: %w", err)
		}
		defer func() { _ = page.Close() }()

		url := fmt.Sprintf("https://chatgpt.com/?model=%s", model)
		if _, err = page.Goto(url); err != nil {
			return "", fmt.Errorf("could not goto: %w", err)
		}
		if err := page.WaitForLoadState(); err != nil {
			return "", fmt.Errorf("could not wait for load state: %w", err)
		}

		time.Sleep(3 * time.Second)

		promptTextAreaSelector := "#prompt-textarea"
		promptTextAreaLocator := page.Locator(promptTextAreaSelector)

		err = promptTextAreaLocator.WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateVisible,
		})
		if err != nil {
			return "", fmt.Errorf(
				"could not wait for prompt text area: %w",
				err,
			)
		}
		if err = promptTextAreaLocator.Fill(prompt); err != nil {
			return "", fmt.Errorf("could not fill prompt text area: %w", err)
		}

		sendPromptButtonSelector := "[data-testid=\"send-button\"]"
		sendPromptButtonLocator := page.Locator(sendPromptButtonSelector)
		if err := sendPromptButtonLocator.Click(); err != nil {
			return "", fmt.Errorf("could not click send prompt button: %w", err)
		}

		feedbackButtonSelector := "[data-testid=\"good-response-turn-action-button\"]"
		feedbackButtonLocator := page.Locator(feedbackButtonSelector)
		err = feedbackButtonLocator.WaitFor(playwright.LocatorWaitForOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(3 * 60 * 1000), // 3 minutes
		})
		if err != nil && !errors.Is(err, playwright.ErrTimeout) {
			return "", fmt.Errorf("could not wait for feedback button: %w", err)
		}

		promptResponseSelector := "[data-message-author-role=\"assistant\"]"
		promptResponseLocator := page.Locator(promptResponseSelector)
		result, err := promptResponseLocator.InnerText()
		if err != nil {
			return "", fmt.Errorf("could not get inner text: %w", err)
		}

		return result, nil
	}
}
