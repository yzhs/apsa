package apsa

// ModernistRecipe describes a recipe with Modernist Cuisine-style steps grouped
// together with ingredients needed for that step.
type ModernistRecipe struct {
	Title    string   `yaml:"title"`
	Portions string   `yaml:"portions"`
	Source   string   `yaml:"source"`
	Tags     []string `yaml:"tags"`
	Steps    []Step   `yaml:"steps"`
}

// Step consisting of ingredients
type Step struct {
	Instructions string   `yaml:"instructions"`
	Ingredients  []string `yaml:"ingredients"`
}
