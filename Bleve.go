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

type Bleve struct {
	Backend Backend
}

func touch(file string) error {
	now := time.Now()
	return os.Chtimes(file, now, now)
}

// BuildIndex generates a new index or updates the documents in an existing one.
func (Bleve) BuildIndex() error {
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
func (b Bleve) SearchBleve(queryString string) (Results, error) {
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
		recipe, err := b.Backend.ReadRecipe(id)
		TryLogError(err)
		ids = append(ids, recipe)
	}

	return Results{ids[:len(searchResults.Hits)], int(searchResults.Total)}, nil
}

// Search return a list of all recipes matching the given query.
func (b Bleve) Search(query string) (Results, error) {
	results, err := b.SearchBleve(query)
	if err != nil {
		return Results{}, err
	}
	n := len(results.Recipes)
	recipes := make([]Recipe, n)
	i := 0
	for _, recipe := range results.Recipes {
		if _, err := os.Stat(Config.KnowledgeDirectory + string(recipe.Id) + ".md"); os.IsNotExist(err) {
			RemoveFromIndex(recipe.Id)
			continue
		}
		recipes[i] = recipe
		i += 1
	}
	results.Total = n // The number of hits can be wrong if recipes have been deleted
	results.Recipes = recipes

	return results, nil
}

func (Bleve) ComputeStatistics() Statistics {
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
