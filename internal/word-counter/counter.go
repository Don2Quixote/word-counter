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
	"sync"

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

	recordsCh, errCh := c.scanSources(ctx, sources, word)

	// We don't need a mutex here to protect records
	// because only one goroutine writes to it and execution flow controlled
	// by done channel.
	records := make([]Record, 0, len(sources))
	done := make(chan struct{})

	handleRecords := func() {
		for record := range recordsCh {
			records = append(records, record)
		}
		done <- struct{}{}
	}
	go handleRecords()

	// errCh will be closed and ok will be false if no error happened.
	err, ok := <-errCh
	if ok {
		return nil, err
	}

	<-done

	return records, nil
}

// scanSources scans sources launching goroutines. Goroutines count controlled by semaphore and configurable
// parameter operationsCountLimit. It returns channel with records and channel with errors. Both channels will
// be closed onces all sources will be scanned.
func (c *Counter) scanSources(ctx context.Context, sources []string, word string) (chan Record, chan error) {
	records := make(chan Record, 1)
	errors := make(chan error, 1)

	wg := &sync.WaitGroup{}
	wg.Add(len(sources))

	go func() {
		for _, source := range sources {
			c.sema.Acquire()
			go func(source string) {
				defer wg.Done()
				defer c.sema.Release()

				count, err := c.countInSource(ctx, source, word)
				if err != nil {
					errors <- fmt.Errorf("count in source %q: %w", source, err)
					return
				}

				records <- Record{
					Source: source,
					Count:  count,
				}
			}(source)
		}
	}()

	go func() {
		wg.Wait()
		close(records)
		close(errors)
	}()

	return records, errors
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
