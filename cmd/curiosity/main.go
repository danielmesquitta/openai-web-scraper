package main

import (
	"context"

	"github.com/danielmesquitta/openai-web-scraper/internal/script/curiosity"
)

func main() {
	generateCuriosities := curiosity.NewGenerateCuriosities()
	if err := generateCuriosities.Run(context.Background()); err != nil {
		panic(err)
	}
}
