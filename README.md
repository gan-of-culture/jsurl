# JSURL
![Github Workflow Status](https://img.shields.io/github/workflow/status/gan-of-culture/jsurl/Go)
[![Go Report Card](https://goreportcard.com/badge/github.com/gan-of-culture/jsurl)](https://goreportcard.com/report/github.com/gan-of-culture/jsurl)
[![Go Reference](https://pkg.go.dev/badge/github.com/gan-of-culture/jsurl.svg)](https://pkg.go.dev/github.com/gan-of-culture/jsurl)

Golang port for [jsurl](https://github.com/Sage/jsurl)

## Example

```golang
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/gan-of-culture/jsurl"
)

type demoStruct struct {
	B interface{}
	C bool
	D int
	E string
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		fmt.Println("you have to supply a jsurl string like ~(B~null~C~false~D~0~E~'hello*20world**203c)")
		return
	}

	if !strings.HasPrefix(args[0], "~(") {
		fmt.Println("jsurl string not found")
		return
	}

	test := &demoStruct{}

	jsurl.Parse(args[0], test)

	jsonData, _ := json.MarshalIndent(*test, "", "    ")
	fmt.Printf("%s\n", jsonData)
}
```

For more examples please take a look at the [unittests](jsurl_test.go)

## License

[GPL-3.0](LICENSE)

