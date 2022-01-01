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
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/blevesearch/bleve"

	"github.com/blevesearch/bleve/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/analysis/analyzer/simple"
)

func touch(file string) error {
	now := time.Now()
	return os.Chtimes(file, now, now)
}

// RebuildIndex generates a new index or updates the documents in an existing one.
func RebuildIndex() error {
	// TODO handle multiple fields, i.e. the main text, @source, @type, tags, etc.
	newIndex := false

	// Try to open an existing index or create a new one if none exists.
	index, err := openIndex()
	if err != nil {
		index = createIndex()
		newIndex = true
	}
	defer index.Close()

	indexUpdateFile := Config.ApsaDirectory + "index_updated"
	indexUpdateTime, err := getModTime(indexUpdateFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Save the time of this indexing operation
			indexUpdateTime = 0
		} else {
			LogError(err)
			return nil
		}
	}

	files, err := ioutil.ReadDir(Config.KnowledgeDirectory)
	if err != nil {
		return err
	}

	batch := index.NewBatch()
	for _, file := range files {
		// Check whether the recipe is newer than the index.
		modTime, err := getModTime(Config.KnowledgeDirectory + file.Name())
		if err != nil {
			LogError(err)
			continue
		}
		id := strings.TrimSuffix(file.Name(), ".md")
		if modTime < indexUpdateTime && !newIndex {
			continue
		}

		// Load and parse the recipe content
		contentBytes, err := ioutil.ReadFile(Config.KnowledgeDirectory + file.Name())
		TryLogError(err)
		content := string(contentBytes)
		recipe := Parse(id, content)

		err = batch.Index(id, recipe)
		TryLogError(err)
	}
	err = index.Batch(batch)
	TryLogError(err)

	// Save the time of this indexing operation
	_ = touch(Config.ApsaDirectory + "index_updated")

	return nil
}

func RemoveFromIndex(id Id) error {
	index, err := openIndex()
	if err != nil {
		return err
	}
	defer index.Close()
	return index.Delete(string(id))
}

// openIndex opens an existing index
func openIndex() (bleve.Index, error) {
	return bleve.Open(Config.ApsaDirectory + "bleve")
}

func createIndex() bleve.Index {
	enTextMapping := bleve.NewTextFieldMapping()
	enTextMapping.Analyzer = "en"

	simpleMapping := bleve.NewTextFieldMapping()
	simpleMapping.Analyzer = simple.Name

	typeMapping := bleve.NewTextFieldMapping()
	typeMapping.Analyzer = keyword.Name

	recipeMapping := bleve.NewDocumentMapping()
	recipeMapping.AddFieldMappingsAt("id", simpleMapping)
	recipeMapping.AddFieldMappingsAt("content", enTextMapping)
	recipeMapping.AddFieldMappingsAt("source", simpleMapping)
	recipeMapping.AddFieldMappingsAt("tag", simpleMapping)

	mapping := bleve.NewIndexMapping()
	mapping.DefaultAnalyzer = "en"
	mapping.DefaultMapping = recipeMapping

	index, err := bleve.New(Config.ApsaDirectory+"bleve", mapping)
	if err != nil {
		panic(err)
	}

	return index
}

// Search the swish index for a given query.
func searchBleve(queryString string) (Results, error) {
	index, err := openIndex()
	if err != nil {
		LogError(err)
		return Results{}, err
	}
	defer index.Close()

	newQueryString := ""
	for _, tmp := range strings.Split(strings.TrimSpace(queryString), " ") {
		word := strings.TrimSpace(tmp)
		if word[0] == '-' || word[0] == '+' {
			newQueryString += " " + word
		} else if word[0] == '~' {
			// Remove prefix to make term optional
			newQueryString += " " + word[1:]
		} else {
			newQueryString += " +" + word
		}
	}

	query := bleve.NewQueryStringQuery(newQueryString[1:]) // Remove leading space
	search := bleve.NewSearchRequest(query)
	search.Size = Config.MaxResults
	searchResults, err := index.Search(search)
	if err != nil {
		println("Invalid query string: '" + newQueryString[1:] + "'")
		LogError(err)
		return Results{}, err
	}

	var ids []Recipe
	for _, match := range searchResults.Hits {
		id := Id(match.ID)
		content, err := readRecipe(id)
		TryLogError(err)
		recipe := Parse(string(id), content)
		ids = append(ids, recipe)
	}

	return Results{ids[:len(searchResults.Hits)], int(searchResults.Total)}, nil
}

// FindRecipes return a list of all recipes matching the given query.
func FindRecipes(query string) (Results, error) {
	results, err := searchBleve(query)
	if err != nil {
		return Results{}, err
	}
	n := ProcessRecipes(results.Ids)
	ids := make([]Recipe, n)
	i := 0
	for _, id := range results.Ids {
		if _, err := os.Stat(Config.KnowledgeDirectory + string(id.Id) + ".md"); os.IsNotExist(err) {
			continue
		}
		ids[i] = Recipe{Id: id.Id}
		i += 1
	}
	results.Total = n // The number of hits can be wrong if recipes have been deleted

	return results, nil
}

func ComputeStatistics() Statistics {
	index, err := openIndex()
	if err != nil {
		LogError(err)
	}
	defer index.Close()

	num, size := getDirSize(Config.KnowledgeDirectory)
	if err == nil {
		tmp, err := index.DocCount()
		if err != nil {
			LogError(err)
		} else {
			num = int(tmp)
		}
	}

	return statistics{num, size}
}
