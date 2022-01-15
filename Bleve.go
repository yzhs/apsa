package apsa

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/blevesearch/bleve"

	"github.com/blevesearch/bleve/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/analysis/analyzer/simple"
	_ "github.com/blevesearch/bleve/analysis/lang/de"
)

type Bleve struct {
	Backend Backend
}

func touch(file string) error {
	now := time.Now()
	return os.Chtimes(file, now, now)
}

// BuildIndex generates a new index or updates the documents in an existing one.
func (b Bleve) BuildIndex() error {
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

		filename := file.Name()
		extension := path.Ext(filename)
		var id string
		if extension == ".md" {
			id = strings.TrimSuffix(filename, ".md")
		} else if extension == ".yaml" {
			id = strings.TrimSuffix(filename, ".yaml")
		} else {
			continue
		}

		if modTime < indexUpdateTime && !newIndex {
			continue
		}

		// Load and parse the recipe content
		recipe, err := b.Backend.ReadRecipe(Id(id))
		TryLogError(err)

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
	textMapping := bleve.NewTextFieldMapping()
	textMapping.Analyzer = "de"

	simpleMapping := bleve.NewTextFieldMapping()
	simpleMapping.Analyzer = simple.Name

	typeMapping := bleve.NewTextFieldMapping()
	typeMapping.Analyzer = keyword.Name

	stepsMapping := bleve.NewDocumentMapping()
	stepsMapping.AddFieldMappingsAt("instructions", textMapping)

	recipeMapping := bleve.NewDocumentMapping()
	recipeMapping.AddFieldMappingsAt("id", simpleMapping)
	recipeMapping.AddFieldMappingsAt("content", textMapping)
	recipeMapping.AddFieldMappingsAt("source", simpleMapping)
	recipeMapping.AddFieldMappingsAt("tags", simpleMapping)
	recipeMapping.AddSubDocumentMapping("steps", stepsMapping)

	mapping := bleve.NewIndexMapping()
	mapping.DefaultAnalyzer = "de"
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

	var recipes []ModernistRecipe
	for _, match := range searchResults.Hits {
		id := Id(match.ID)
		recipe, err := b.Backend.ReadRecipe(id)
		if err != nil {
			log.Println("Error reading recipe:", err)
			recipe.Id = id
		}

		recipes = append(recipes, recipe)
	}

	return Results{recipes[:len(searchResults.Hits)], int(searchResults.Total)}, nil
}

// Search return a list of all recipes matching the given query.
func (b Bleve) Search(query string) (Results, error) {
	results, err := b.SearchBleve(query)
	if err != nil {
		return Results{}, err
	}
	n := len(results.Recipes)
	recipes := make([]ModernistRecipe, n)
	i := 0
	for _, recipe := range results.Recipes {
		if !b.Backend.RecipeExists(recipe.Id) {
			log.Printf("Could not find file for recipe %s. Removing it from the index.\n", recipe.Id)
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
