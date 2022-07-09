package wordcounter

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"word-counter/pkg/sema"
)

// Counter is general logic entity for counting words in specified sources.
type Counter struct {
	// sema is semaphore for limiting operations count.
	sema *sema.Sema
	// httpClient is client for reading sources represented as web urls.
	httpClient http.Client
}

// NewCounter returns new initialized counter.
func NewCounter(operationsCountLimit int, httpClient http.Client) *Counter {
	return &Counter{
		sema:       sema.New(operationsCountLimit),
		httpClient: httpClient,
	}
}

// Record is struct with Count of words for a particular Source.
type Record struct {
	Source string
	Count  int
}

// re is regexp for searching words.
var re = regexp.MustCompile(`[\w-]+`)

// Count counts word in sources and returns records for each source.
func (c *Counter) Count(ctx context.Context, sources []string, word string) ([]Record, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	recordsCh := make(chan Record)
	errCh := make(chan error)

	for _, source := range sources {
		go func(source string) {
			c.sema.Acquire()
			defer c.sema.Release()

			count, err := c.countInSource(ctx, source, word)
			if err != nil {
				errCh <- fmt.Errorf("count in source %q: %w", source, err)
			}

			recordsCh <- Record{
				Source: source,
				Count:  count,
			}
		}(source)
	}

	records := make([]Record, 0, len(sources))

	for len(records) != len(sources) {
		select {
		case record := <-recordsCh:
			records = append(records, record)
		case err := <-errCh:
			cancel()
			return nil, err
		}
	}

	return records, nil
}

// countIsSource returns count of word in source.
func (c *Counter) countInSource(ctx context.Context, source string, word string) (int, error) {
	content, err := c.read(ctx, source)
	if err != nil {
		return 0, fmt.Errorf("read source: %w", err)
	}

	words := re.FindAllString(string(content), -1)
	count := 0

	for _, w := range words {
		if strings.EqualFold(w, word) {
			count++
		}
	}

	return count, nil
}

// read reads content from source. Source can be either path to file or web url.
func (c *Counter) read(ctx context.Context, source string) ([]byte, error) {
	// os.Stat returns nil error if file exists.
	_, err := os.Stat(source)
	if err == nil {
		return os.ReadFile(source)
	}

	_, err = url.ParseRequestURI(source)
	if err != nil {
		return nil, fmt.Errorf("invalid source")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status is not OK (%s)", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	return body, nil
}
