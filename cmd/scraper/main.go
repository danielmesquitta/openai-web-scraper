package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/danielmesquitta/openai-web-scraper/internal/pkg/jsonutil"
	"github.com/playwright-community/playwright-go"
)

func main() {
	runPlaywright()
}

const tmpFolderPath = "tmp"

const templatePrompt = "Write 100 objects for a title and interesting fact about the %s breed, and return in JSON format as pure text, not as code, with the keys \"title\" and \"content\". Be direct and return only the JSON."

const breed = "Shih Tzu"

func runPlaywright() {
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	browser, err := pw.Chromium.ConnectOverCDP("http://localhost:9222")
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	contexts := browser.Contexts()
	if len(contexts) == 0 {
		log.Fatalf("could not get contexts")
	}
	page, err := contexts[0].NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}
	if _, err = page.Goto("https://chatgpt.com/?model=o1-preview"); err != nil {
		log.Fatalf("could not goto: %v", err)
	}
	if err := page.WaitForLoadState(); err != nil {
		log.Fatalf("could not wait for load state: %v", err)
	}

	time.Sleep(3 * time.Second)

	promptTextAreaSelector := "#prompt-textarea"
	promptTextAreaLocator := page.Locator(promptTextAreaSelector)

	err = promptTextAreaLocator.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		log.Fatalf("could not wait for prompt text area: %v", err)
	}
	prompt := fmt.Sprintf(templatePrompt, breed)
	if err = promptTextAreaLocator.Fill(prompt); err != nil {
		log.Fatalf("could not fill prompt text area: %v", err)
	}

	sendPromptButtonSelector := "[data-testid=\"send-button\"]"
	sendPromptButtonLocator := page.Locator(sendPromptButtonSelector)
	if err := sendPromptButtonLocator.Click(); err != nil {
		log.Fatalf("could not click send prompt button: %v", err)
	}

	promptResponseSelector := "[data-message-author-role=\"assistant\"]"
	promptResponseLocator := page.Locator(promptResponseSelector)
	if err := promptResponseLocator.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(3 * 60 * 1000), // 3 minutes
	}); err != nil {
		log.Fatalf("could not wait for prompt response: %v", err)
	}

	feedbackButtonSelector := "[data-testid=\"good-response-turn-action-button\"]"
	feedbackButtonLocator := page.Locator(feedbackButtonSelector)
	if err := feedbackButtonLocator.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5 * 60 * 1000), // 5 minutes
	}); err != nil {
		log.Fatalf("could not wait for feedback button: %v", err)
	}

	innerText, err := promptResponseLocator.InnerText()
	if err != nil {
		log.Fatalf("could not get inner text: %v", err)
	}

	jsonData := jsonutil.ExtractJSONFromText(innerText)
	for _, jsonItem := range jsonData {
		fileName := formatTextToFileName(breed)
		filePath := fmt.Sprintf(
			"%s/%s.json",
			tmpFolderPath,
			fileName,
		)
		if err := os.WriteFile(filePath, []byte(jsonItem), 0644); err != nil {
			log.Fatalf("could not write to file: %v", err)
		}
	}
}

func formatTextToFileName(text string) string {
	return strings.ReplaceAll(strings.ToLower(text), " ", "-")
}
