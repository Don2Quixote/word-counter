package app

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	wordcounter "word-counter/internal/word-counter"
)

// Run runs app. If returned error is not nil, program exited
// unexpectedly and non-zero code should be returned (os.Exit(1) or log.Fatal(...)).
func Run(ctx context.Context) error {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}

	sources := strings.Split(string(input), "\n")
	if sources[len(sources)-1] == "" {
		sources = sources[:len(sources)-1]
	}

	counter := wordcounter.NewCounter(_operationsCountLimit, http.Client{
		Timeout: _httpTimeout,
	})

	records, err := counter.Count(ctx, sources, _wordToCount)
	if err != nil {
		return fmt.Errorf("count: %w", err)
	}

	total := 0

	for _, record := range records {
		log.Printf("Count for %s: %d\n", record.Source, record.Count)
		total += record.Count
	}

	log.Printf("Total: %d", total)

	return nil
}
