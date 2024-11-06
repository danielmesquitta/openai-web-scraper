package curiosity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path"
	"strings"

	"github.com/danielmesquitta/openai-web-scraper/internal/config"
	"github.com/danielmesquitta/openai-web-scraper/internal/pkg/jsonutil"
	"github.com/danielmesquitta/openai-web-scraper/internal/provider/scraper"
)

const wantedCuriosities = 730 // 2 years of curiosities

const expectedCuriositiesPerLoop = 100

const templatePrompt = "Write %d objects for a title and interesting fact about the %s breed," +
	"and return an array of objects (with keys \"title\" and \"content\") in JSON format as pure text, not as code."

type GenerateCuriosities struct {
	e *config.Env
	s scraper.Scraper
}

func newGenerateCuriosities(
	e *config.Env,
	s scraper.Scraper,
) *GenerateCuriosities {
	return &GenerateCuriosities{
		e: e,
		s: s,
	}
}

type Curiosity struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (gc *GenerateCuriosities) Run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return errors.New("context canceled or deadline exceeded")
	default:
		for _, breed := range gc.e.Breeds {
			if err := gc.run(ctx, breed); err != nil {
				log.Printf("could not run for breed %s: %v", breed, err)
			}
		}
	}

	return nil
}

func (gc *GenerateCuriosities) run(
	ctx context.Context,
	breed string,
) error {
	select {
	case <-ctx.Done():
		return errors.New("context canceled or deadline exceeded")
	default:
		fileName := fmt.Sprintf("%s.json", breed)
		filePath := path.Join(gc.e.CuriositiesFolderPath, fileName)
		basePrompt := fmt.Sprintf(
			templatePrompt,
			expectedCuriositiesPerLoop,
			breed,
		)
		initialCuriosities, err := gc.getCuriositiesFromJSONFile(filePath)
		if err != nil {
			return fmt.Errorf(
				"could not get initial curiosities from json file: %w",
				err,
			)
		}

		currentCuriosities := float64(len(initialCuriosities))

		missingCuriosities := wantedCuriosities - currentCuriosities
		if missingCuriosities <= 0 {
			return nil
		}

		loopsCount := math.Ceil(
			missingCuriosities / expectedCuriositiesPerLoop,
		)

		for i := range int(loopsCount) {
			if err := gc.generateCuriosities(ctx, filePath, basePrompt); err != nil {
				fmt.Printf(
					"could not generate curiosities for breed %s attempt %d: %v",
					breed,
					i,
					err,
				)
				continue
			}
		}

		return nil
	}
}

func (gc *GenerateCuriosities) generateCuriosities(
	ctx context.Context,
	filePath,
	basePrompt string,
) error {
	curiosities, err := gc.getCuriositiesFromJSONFile(filePath)
	if err != nil {
		return fmt.Errorf(
			"could not get curiosities from json file: %w",
			err,
		)
	}

	if len(curiosities) >= wantedCuriosities {
		return nil
	}

	prompt := basePrompt
	if len(curiosities) > 0 {
		jsonCuriosities, err := json.Marshal(curiosities)
		if err != nil {
			return fmt.Errorf("could not marshal json: %w", err)
		}
		prompt += fmt.Sprintf(
			"Do not repeat the same topics from the following JSON: %s",
			jsonCuriosities,
		)
	}

	result, err := gc.s.ScrapOpenAIPrompt(
		ctx,
		prompt,
		scraper.OpenAIModelO1Preview,
	)
	if err != nil {
		return fmt.Errorf("could not scrap openai prompt: %w", err)
	}

	jsonResults := jsonutil.ExtractJSONFromText(result)
	if len(jsonResults) == 0 {
		return fmt.Errorf("no json data found in result")
	}

	jsonResult := jsonResults[0]
	if len(jsonResults) > 1 {
		jsonResult = fmt.Sprintf("[%s]", strings.Join(jsonResults, ","))
	}

	newCuriosities := []Curiosity{}
	if err := json.Unmarshal([]byte(jsonResult), &newCuriosities); err != nil {
		return fmt.Errorf("could not unmarshal json %s: %w", jsonResult, err)
	}

	curiosities = append(curiosities, newCuriosities...)

	uniqueCuriosities := gc.getUniqueCuriosities(curiosities)

	if err := gc.updateJSONFileCuriosities(filePath, uniqueCuriosities); err != nil {
		return fmt.Errorf(
			"could not update json file curiosities: %w",
			err,
		)
	}

	return nil
}

func (gc *GenerateCuriosities) getCuriositiesFromJSONFile(
	filePath string,
) ([]Curiosity, error) {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Curiosity{}, nil
		}

		return []Curiosity{}, fmt.Errorf("could not open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return []Curiosity{}, fmt.Errorf("could not read file: %w", err)
	}

	var curiosities []Curiosity
	if err := json.Unmarshal(bytes, &curiosities); err != nil {
		return []Curiosity{}, fmt.Errorf("could not unmarshal json: %w", err)
	}

	return curiosities, nil
}

func (gc *GenerateCuriosities) getUniqueCuriosities(
	curiosities []Curiosity,
) []Curiosity {
	uniqueCuriosities := make([]Curiosity, len(curiosities))
	copy(uniqueCuriosities, curiosities)

	uniqueCuriositiesByTitle := map[string]Curiosity{}
	for _, curiosity := range uniqueCuriosities {
		uniqueCuriositiesByTitle[curiosity.Title] = curiosity
	}

	uniqueCuriosities = []Curiosity{}
	for _, curiosity := range uniqueCuriositiesByTitle {
		uniqueCuriosities = append(uniqueCuriosities, curiosity)
	}

	uniqueCuriositiesByContent := map[string]Curiosity{}
	for _, curiosity := range uniqueCuriosities {
		uniqueCuriositiesByContent[curiosity.Content] = curiosity
	}

	uniqueCuriosities = []Curiosity{}
	for _, curiosity := range uniqueCuriositiesByContent {
		uniqueCuriosities = append(uniqueCuriosities, curiosity)
	}

	return uniqueCuriosities
}

func (gc *GenerateCuriosities) updateJSONFileCuriosities(
	filePath string,
	curiosities []Curiosity,
) error {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("could not open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	bytes, err := json.Marshal(curiosities)
	if err != nil {
		return fmt.Errorf("could not marshal json: %w", err)
	}

	if _, err := file.Write(bytes); err != nil {
		return fmt.Errorf("could not write to file: %w", err)
	}

	return nil
}
