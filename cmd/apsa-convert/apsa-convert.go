// Package converter contains a converter from old-style to new style YAML-based
// recipe files.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/yzhs/apsa"
)

func main() {
	apsa.InitConfig()

	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	entries, err := os.ReadDir(fmt.Sprintf("%s/.apsa/library", home))
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		file := entry.Name()
		if strings.HasSuffix(file, ".md") {
			fmt.Println(file)
		}
	}
}
