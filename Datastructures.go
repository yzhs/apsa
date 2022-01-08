package apsa

import (
	"html/template"
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

type Backend interface {
	ReadRecipe(id Id) (Recipe, error)
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

type Results struct {
	// Recipes contains the recipes to be displayed.
	Recipes []Recipe

	// Total number of results there were all in all; can be significantly
	// larger than the number of Recipes
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

func NewSearchEngine() SearchEngine {
	return Bleve{MarkdownParser{}}
}
