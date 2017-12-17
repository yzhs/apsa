// Apsa
//
// Copyright (C) 2015-2016  Colin Benner
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package apsa

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/russross/blackfriday"
)

var ErrNoSuchRecipe = errors.New("No such recipe")

const hashes = "############################################################"

type errTemplateReader struct {
	doc string
	err error
}

// Load a template file from disk and propagate errors
func (e *errTemplateReader) readTemplate(name string) {
	if e.err != nil {
		return
	}

	tmp, err := readTemplate(name)
	e.err = err
	e.doc += tmp
}

func writeRecipe(id Id, html []byte) error {
	return ioutil.WriteFile(Config.CacheDirectory+string(id)+".html", html, 0644)
}

func runPandoc(id Id) error {
	f, err := os.Open(Config.KnowledgeDirectory + string(id) + ".md")
	if os.IsNotExist(err) {
		return RemoveFromIndex(id)
	} else if err != nil {
		panic(err)
	}
	f.Close()

	recipeString, err := readRecipe(id)
	if err != nil {
		panic(err)
	}
	recipe := Parse(string(id), recipeString)

	output := blackfriday.MarkdownCommon([]byte(recipe.Content))
	return writeRecipe(id, output)
}

func recipeToHTMLSnippet(id Id) error {
	runPandoc(id)
	// TODO generate HTML file
	return nil
}

// Generate a HTML snippet from a given recipe whenever necessary.
func ProcessRecipe(id Id) error {
	if isUpToDate(id) {
		return nil
	}

	return recipeToHTMLSnippet(id)
}

// Generate images for a list of recipes.
func ProcessRecipes(ids []Recipe) int {
	numRecipes := 0

	for _, foo := range ids {
		id := foo.Id
		err := ProcessRecipe(id)
		if err == ErrNoSuchRecipe {
			continue
		} else if err != nil {
			log.Panic("An error ocurred when processing recipe ", id, ": ", err)
		} else {
			numRecipes++
		}
	}

	return numRecipes
}

func RenderAllRecipes() int {
	files, err := ioutil.ReadDir(Config.KnowledgeDirectory)
	if err != nil {
		panic(err)
	}
	var errors []error
	limitGoroutines := make(chan bool, Config.MaxProcs)
	for i := 0; i < Config.MaxProcs; i++ {
		limitGoroutines <- true
	}
	ch := make(chan int, len(files))
	for _, file := range files {
		go func(file os.FileInfo) {
			<-limitGoroutines
			if !strings.HasSuffix(file.Name(), ".md") {
				ch <- 0
				return
			}
			id := Id(strings.TrimSuffix(file.Name(), ".md"))
			if err := ProcessRecipe(id); err != nil && err != ErrNoSuchRecipe {
				log.Printf("%s\nERROR\n%s\n%v\n%s\n", hashes, hashes, err, hashes)
			}
			ch <- 1
		}(file)
	}
	counter := 0
	for i := 0; i < len(files); i++ {
		counter += <-ch
		limitGoroutines <- true
	}
	for _, err := range errors {
		log.Printf("Error: %v\n", err)
	}
	return counter
}
