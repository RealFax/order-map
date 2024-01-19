# Order map - Keep everything in order

_We ensure that the API and semantics are compatible with `sync.Map`. If you have any questions, please raise them in the issue_

## Installation
`go get github.com/RealFax/order-map@latest`

## Quickstart

```go
package main

import odmap "github.com/RealFax/order-map"

func main() {
	m := odmap.New[int, string]()
	m.Store(0, "Hello")
	m.Store(1, "World")
	m.Store(2, "😄😄😄")

	m.Range(func(key int, value string) bool {
		print(value, " ")
		return true
	})
}
```
## Roadmap
_Welcome to propose more features in the issue_

- [x] Concurrency safety (add `--tags=safety_map` enabled)

_⚠️Note. Features such as: Len, Contains are not stable and may be removed or have semantic changes in the future._