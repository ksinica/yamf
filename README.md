# yamf

[![GoDoc](https://godoc.org/github.com/ksinica/yamf?status.svg)](https://godoc.org/github.com/ksinica/yamf)

[Yet-Another-Multi-Format](https://github.com/AljoschaMeyer/yamf-hash) implementation in Go.

#### Warning
This library is a work in progress and does not strive to have an A+ score at the moment. Although, `v0.1.0` should be simple, usable, and self-explanatory enough with fair unit test coverage.

#### Example
##### Encode hash to JSON
```go
package main

import (
	"encoding/json"
	"os"

	"github.com/ksinica/yamf"
	"github.com/ksinica/yamf/hash/blake2b"
)

func main() {
	hash, _ := blake2b.New()
	hash.Write([]byte("Hello world!"))

	json.NewEncoder(os.Stdout).Encode(struct {
		Hash *yamf.TypeValue `json:"hash"`
	}{
		Hash: yamf.HashToTypeValue(hash),
	})
	// {"hash":"0008239dad53cf69f358"}
}
```
##### Decode hash from JSON
```go
package main

import (
	"encoding/json"

	"github.com/ksinica/yamf"
	_ "github.com/ksinica/yamf/hash/blake2b"
)

func main() {
	var val struct {
		Hash *yamf.TypeValue `json:"hash"`
	}

	json.Unmarshal([]byte(`{"hash":"0008239dad53cf69f358"}`), &val)

	hash, _ := yamf.TypeValueToHash(*val.Hash)
	hash.Write([]byte("Hello world!"))

	if !yamf.HashEqual(hash, *val.Hash) {
		panic("oh no!")
	}
}
```
## License

Source code is available under the MIT [License](/LICENSE).
