package apsa

import (
	"strings"

	"gopkg.in/yaml.v2"
)

// ModernistRecipe describes a recipe with Modernist Cuisine-style steps grouped
// together with ingredients needed for that step.
type ModernistRecipe struct {
	Id       Id
	Title    string   `yaml:"title"`
	Portions string   `yaml:"portions"`
	Source   string   `yaml:"source"`
	Tags     []string `yaml:"tags"`
	Steps    []Step   `yaml:"steps"`
}

// Step consisting of ingredients
type Step struct {
	Instructions []string `yaml:"instructions"`
	Ingredients  []string `yaml:"ingredients"`
}

func FromRecipe(recipe Recipe) ModernistRecipe {
	return ModernistRecipe{
		Id:       recipe.Id,
		Title:    recipe.Title,
		Portions: recipe.Portions,
		Source:   recipe.Source,
		Tags:     recipe.Tags,
		Steps: []Step{
			{
				Ingredients:  recipe.Ingredients,
				Instructions: strings.Split(recipe.Content, "\n\n"),
			},
		},
	}
}

type YamlParser struct {
	fileReader FileReader
}

func (y YamlParser) ReadRecipe(id Id) (ModernistRecipe, error) {
	content, err := y.readRecipe(id)
	TryLogError(err)
	recipe := y.Parse(id, content)
	return recipe, err
}

// Load the content of a given recipe from disk.
func (y YamlParser) readRecipe(id Id) ([]byte, error) {
	return y.fileReader.ReadFile(Config.KnowledgeDirectory + string(id) + ".yaml")
}

func (YamlParser) Parse(id Id, doc []byte) ModernistRecipe {
	var recipe ModernistRecipe
	yaml.Unmarshal(doc, recipe)
	return recipe
}
