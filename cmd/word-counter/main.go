package main

import (
	"context"
	"fmt"
	"log"

	"word-counter/pkg/graceful"

	"word-counter/internal/app"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	graceful.OnShutdown(cancel)

	err := app.Run(ctx)
	if err != nil {
		log.Fatal(fmt.Errorf("error running app: %w", err))
	}
}
