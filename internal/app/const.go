package app

import "time"

// _wordToCount is word to count in sources.
const _wordToCount = "Go"

// _operationsCountLimit is value for limiting maximum number
// of concurrently executing counting operations.
const _operationsCountLimit = 5

// _httpTimeout is timeout for constructing http.Client.
const _httpTimeout = time.Second * 30
