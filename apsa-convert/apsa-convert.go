// Package converter contains a converter from old-style to new style YAML-based
// recipe files.
package main

import (
	"fmt"
	"os"

	. "github.com/yzhs/apsa"
)

func main() {
	InitConfig()

	for _, arg := range os.Args {
		file := arg
		fmt.Println(file)
	}
}
