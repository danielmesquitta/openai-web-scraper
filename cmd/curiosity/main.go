package main

import (
	"context"
	"log"
	"os/signal"
	"sync"
	"syscall"

	"github.com/danielmesquitta/openai-web-scraper/internal/script/curiosity"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		generateCuriosities := curiosity.NewGenerateCuriosities()
		if err := generateCuriosities.Run(ctx); err != nil {
			panic(err)
		}
	}()

	<-ctx.Done()
	stop()

	log.Println("Shutting down...")

	wg.Wait()
}
