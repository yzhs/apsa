package apsa

import (
	"os"
)

const (
	NAME    = "Apsa"
	VERSION = "0.1"
)

// statistics concerning the size of the library.
type statistics struct {
	numRecipes int
	fileSize   int64
}

func (s statistics) Num() int {
	return s.numRecipes
}

func (s statistics) Size() int64 {
	return s.fileSize
}

type Id string

type SearchEngine interface {
	BuildIndex() error
	Search(query string) (Results, error)
	ComputeStatistics() Statistics
}

type Renderer interface {
	Extension() string
	Render(id Id) error
}

type Statistics interface {
	Num() int
	Size() int64
}

// Configuration data of Apsa
type Configuration struct {
	// How many processes may run in parallel when rendering
	MaxProcs int

	// How many results are to be processed at once
	MaxResults int

	ApsaDirectory      string
	KnowledgeDirectory string
	TemplateDirectory  string
	TempDirectory      string
}

type Results struct {
	Ids []Recipe
	// How many results there were all in all; can be significantly larger than len(Ids).
	Total int
}

var Config Configuration

func InitConfig() {
	Config.MaxResults = 1000
	Config.MaxProcs = 4

	dir := os.Getenv("HOME") + "/.apsa/"

	Config.ApsaDirectory = dir
	Config.KnowledgeDirectory = dir + "library/"
	Config.TemplateDirectory = dir + "templates/"
	Config.TempDirectory = dir + "tmp/"
}

func CreateSearchEngine() SearchEngine {
	return Bleve{}
}
