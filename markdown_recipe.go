package apsa

import (
	"strings"
)

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

type MarkdownParser struct{}

func (p MarkdownParser) ReadRecipe(id Id) (Recipe, error) {
	content, err := readRecipe(id)
	TryLogError(err)
	recipe := p.Parse(string(id), content)
	return recipe, err
}

// Parse the tags in the given recipe content.  The format of a recipe is
// generally of the following form:
//
//	# Titel
//	Quelle: http://...
//	Backzeit: 23 min
//	Wartezeit: 15 min
//	Zubereitungszeit: 2h + 5min
//	Ober- und Unterhitze: 200
//	Umluft: 180
//	Tags: Gemüse, lecker, gesund
//	Portionen: 4
//	## Teilrezept 1
//	Zutaten:
//	* 1 Rotkohl
//	* 2 EL Zucker
//	* Salz
//	Zubereitung...
//
//	## Teilrezept 2
//	* Rosinen
//	* Schokopudding
//	Zubereitung...
//
// oder
//	[...]
//	Portionen: 4
//	Zutaten:
//	* Wasser
//	Zubereitung...
func (MarkdownParser) Parse(id, doc string) Recipe {
	lines := strings.Split(doc, "\n")
	title, metadata, otherLines := extractMetadata(lines)

	for len(otherLines) > 0 && (strings.TrimSpace(otherLines[0]) == "" || strings.Contains(otherLines[0], "Zutaten:")) {
		otherLines = otherLines[1:]
	}

	ingredients, lastIngredientLine := extractIngredients(otherLines)

	instructions := ""
	if lastIngredientLine < len(otherLines)-1 {
		instructions = strings.Join(otherLines[lastIngredientLine+1:], "\n")
	}

	return Recipe{
		Id:       Id(id),
		Content:  instructions,
		Title:    title,
		Portions: metadata["Portionen"],
		Source:   metadata["Quelle"],
		Tags:     parseTags(metadata["Tags"]),

		CookingTime:     metadata["Kochzeit"],
		BakingTime:      metadata["Backzeit"],
		WaitingTime:     metadata["Wartezeit"],
		TotalTime:       metadata["Gesamtzeit"],
		PreparationTime: metadata["Zubereitungszeit"],

		FanTemp:              metadata["Umfluft"],
		TopAndBottomHeatTemp: metadata["Ober- und Unterhitze"],
		Ingredients:          ingredients,
	}
}

func extractMetadata(lines []string) (string, map[string]string, []string) {
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
	return title, data, otherLines
}

func extractIngredients(otherLines []string) ([]string, int) {
	ingredients := make([]string, 0, 10)
	lastIngredientLine := 0
	for i, line_ := range otherLines {
		line := strings.TrimSpace(line_)
		if strings.HasPrefix(line, "##") {
			// Multipart recipes are not handled properly at the moment
			return make([]string, 0), -1
		}

		if strings.HasPrefix(line, "* ") {
			ingredients = append(ingredients, line[2:])
			lastIngredientLine = i
		}
	}
	return ingredients, lastIngredientLine
}
