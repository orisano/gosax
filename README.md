# gosax

[![Go Reference](https://pkg.go.dev/badge/github.com/orisano/gosax.svg)](https://pkg.go.dev/github.com/orisano/gosax)

`gosax` is a Go library for XML SAX (Simple API for XML) parsing, supporting read-only functionality. This library is
designed for efficient and memory-conscious XML parsing, drawing inspiration from various sources to provide a
performant parser.

## Features

- **Read-only SAX parsing**: Stream and process XML documents without loading the entire document into memory.
- **Efficient parsing**: Utilizes techniques inspired by `quick-xml` and `pkg/json` for high performance.
- **SWAR (SIMD Within A Register)**: Optimizations for fast text processing, inspired by `memchr`.
- **Compatibility with encoding/xml**: Includes utility functions to bridge `gosax` types with `encoding/xml` types, facilitating easy integration with existing code that uses the standard library.

## Benchmark
```
goos: darwin
goarch: arm64
pkg: github.com/orisano/gosax
BenchmarkReader_Event-12    	       5	 211845800 ns/op	1103.30 MB/s	 2097606 B/op	       6 allocs/op
```

## Installation

To install `gosax`, use `go get`:

```bash
go get github.com/orisano/gosax
```

## Usage

Here is a basic example of how to use `gosax` to parse an XML document:

```go
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/orisano/gosax"
)

func main() {
	xmlData := `<root><element>Value</element></root>`
	reader := strings.NewReader(xmlData)

	r := gosax.NewReader(reader)
	for {
		e, err := r.Event()
		if err != nil {
			log.Fatal(err)
		}
		if e.Type() == gosax.EventEOF {
			break
		}
		fmt.Println(string(e.Bytes))
	}
	// Output:
	// <root>
	// <element>
	// Value
	// </element>
	// </root>
}

```

### Bridging with encoding/xml

**Important Note for encoding/xml Users:**
> When migrating from `encoding/xml` to `gosax`, note that self-closing tags are handled differently. To mimic `encoding/xml` behavior, set `gosax.Reader.EmitSelfClosingTag` to `true`. This ensures self-closing tags are recognized and processed correctly.

#### Using TokenE
If you are used to `encoding/xml`'s `Token`, start with `gosax.TokenE`. 
**Note:** Using `gosax.TokenE` and `gosax.Token` involves memory allocation due to interfaces.

**Before:**
```go
var dec *xml.Decoder
for {
	tok, err := dec.Token()
	if err == io.EOF {
		break
	}
	// ...
}
```

**After:**
```go
var dec *gosax.Reader
for {
	tok, err := gosax.TokenE(dec.Event())
	if err == io.EOF {
		break
	}
	// ...
}
```

#### Utilizing xmlb
`xmlb` is an extension for `gosax` to simplify rewriting code from `encoding/xml`. It provides a higher-performance bridge for XML parsing and processing.

**Before:**
```go
var dec *xml.Decoder
for {
	tok, err := dec.Token()
	if err == io.EOF {
		break
	}
	switch t := tok.(type) {
	case xml.StartElement:
		// ...
	case xml.CharData:
		// ...
	case xml.EndElement:
		// ...
	}
} 
```

**After:**
```go
var dec *xmlb.Decoder
for {
	tok, err := dec.Token()
	if err == io.EOF {
		break
	}
	switch tok.Type() {
	case xmlb.StartElement:
		t, _ := tok.StartElement()
		// ...
	case xmlb.CharData:
		t, _ := tok.CharData()
		// ...
	case xmlb.EndElement:
		t := tok.EndElement()
		// ...
	}
} 
```

## License

This library is licensed under the terms specified in the LICENSE file.

## Acknowledgements

`gosax` is inspired by the following projects and resources:

- [Dave Cheney's GopherCon SG 2023 Talk](https://dave.cheney.net/paste/gophercon-sg-2023.html)
- [quick-xml](https://github.com/tafia/quick-xml)
- [memchr](https://github.com/BurntSushi/memchr) (SWAR part)

## Contributing

Contributions are welcome! Please fork the repository and submit pull requests.

## Contact

For any questions or feedback, feel free to open an issue on the GitHub repository.
