// Package converter contains a converter from old-style to new style YAML-based
// recipe files.
package main

import (
	"fmt"
	"io/ioutil"
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

	libraryDir := fmt.Sprintf("%s/.apsa/library", home)
	entries, err := os.ReadDir(libraryDir)
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		file := entry.Name()
		if !strings.HasSuffix(file, ".md") {
			continue
		}
		id := file[:len(file)-3]
		fmt.Printf("Parsing recipe %s...\n", id)
		fileContent, err := ioutil.ReadFile(libraryDir + "/" + file)
		if err != nil {
			panic(err)
		}
		recipe := apsa.Parse(file[:len(file)-3], string(fileContent))
		fmt.Println(recipe.Title)
	}
}
