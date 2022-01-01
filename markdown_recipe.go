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
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/russross/blackfriday"
)

type Recipe struct {
	Id                   Id            `json:"id"`
	Title                string        `json:"titel"`
	Source               string        `json:"quelle"`
	PreparationTime      string        `json:"zubereitungszeit"`
	BakingTime           string        `json:"backzeit"`
	CookingTime          string        `json:"kochzeit"`
	WaitingTime          string        `json:"wartezeit"`
	TotalTime            string        `json:"gesamtzeit"`
	FanTemp              string        `json:"umluft"`
	TopAndBottomHeatTemp string        `json:"oberuntunterhitze"`
	Ingredients          []string      `json:"zutaten"`
	Portions             string        `json:"portionen"`
	Content              string        `json:"inhalt"`
	Tags                 []string      `json:"tag"`
	HTML                 template.HTML `json:""`
}

// Parsing

// Parse a comma separated list of tags into a slice.
func parseTags(line string) []string {
	var tags []string
	for _, tag := range strings.Split(line, ",") {
		tmp := strings.TrimSpace(tag)
		if tmp != "" {
			tags = append(tags, tmp)
		}
	}
	return tags
}

func hasPrefix(needle string, haystack string) bool {
	return strings.HasPrefix(strings.ToLower(needle), strings.ToLower(haystack))
}

// Parse the tags in the given recipe content.  The format of a recipe is
// generally of the following form:
//
//	# Titel
//	Sprache: de
//	Quelle: http://...
//	Backzeit: 23 min
//	Wartezeit: 15 min
//	Zubereitungszeit: 2h + 5min
//	Temperatur:
//		- Ober- und Unterhitze: 200
//		- Umluft: 180
//	Tags: Gemüse, lecker, gesund
//	Portionen: 4
//	## Teilrezept 1
//	Zutaten:
//		- 1 Rotkohl
//		- 2 EL Zucker
//		- Salz
//	Zubereitung...
//
//	## Teilrezept 2
//		- Rosinen
//		- Schokopudding
//	Zubereitung...
//
// oder
//	[...]
//	Portionen: 4
//	Zutaten:
//		- Wasser
//	Zubereitung...
func Parse(id, doc string) Recipe {
	lines := strings.Split(doc, "\n")
	var otherLines []string
	data := make(map[string]string)

	title := strings.TrimSpace(strings.TrimPrefix(lines[0], "#"))
	for _, line_ := range lines[1:] {
		line := strings.TrimSpace(line_)
		containsMetadata := false
		if strings.Contains(line, ":") {
			lst := strings.SplitN(line, ":", 2)
			prefix := strings.ToLower(strings.TrimSpace(lst[0]))
			remainder := strings.TrimSpace(lst[1])
			if prefix == "zutaten" && remainder == "" {
				continue
			}
			metadataTypes := []string{
				"Quelle", "Tags", "Portionen",
				"Zubereitungszeit", "Kochzeit", "Backzeit", "Wartezeit",
				"Gesamtzeit", "Umluft", "Ober- und Unterhitze",
			}
			for _, typ := range metadataTypes {
				if prefix == strings.ToLower(typ) {
					containsMetadata = true
					data[typ] += remainder
					break
				}
			}
		}
		if !containsMetadata {
			otherLines = append(otherLines, line)
		}
	}
	content := strings.Join(otherLines, "\n")

	return Recipe{
		Id:       Id(id),
		Content:  content,
		Title:    title,
		Portions: data["Portionen"],
		Source:   data["Quelle"],
		Tags:     parseTags(data["Tags"]),

		CookingTime:     data["Kochzeit"],
		BakingTime:      data["Backzeit"],
		WaitingTime:     data["Wartezeit"],
		TotalTime:       data["Gesamtzeit"],
		PreparationTime: data["Zubereitungszeit"],

		FanTemp:              data["Umfluft"],
		TopAndBottomHeatTemp: data["Ober- und Unterhitze"],
		// TODO ingredients
	}
}

// Rendering

var ErrNoSuchRecipe = errors.New("no such recipe found")

const hashes = "############################################################"

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
	return runPandoc(id)
}

// ProcessRecipe generates a HTML snippet from a given recipe whenever necessary.
func ProcessRecipe(id Id) error {
	if isUpToDate(id) {
		return nil
	}

	return recipeToHTMLSnippet(id)
}

// ProcessRecipes generates HTML for a list of recipes.
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
