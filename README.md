# word-counter

Reads list of sources from stdin and finds word "Go" (specified in `internal/app/const.go` as `_wordToFind`) there.
Source can be represented either as path to file or as a web url.

Concurrent counting operations are limited with semaphore (count of operations is specified in `internal/app/const.go` as `_operationsCountLimit`).

## Usage example
```
$ go build ./cmd/word-counter
$ echo -e 'https://golang.org\n./internal/app/const.go\nhttps://golang.org\nhttps://golang.org' | ./word-counter
2022/07/08 14:08:41 Count for ./internal/app/const.go: 1
2022/07/08 14:08:41 Count for https://golang.org: 51
2022/07/08 14:08:41 Count for https://golang.org: 51
2022/07/08 14:08:41 Count for https://golang.org: 51
2022/07/08 14:08:41 Total: 154
```