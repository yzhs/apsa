package apsa

import (
	"html/template"
	"strings"
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
	title, metadata, otherLines := extractMetadata(lines)

	content := strings.Join(otherLines, "\n")

	return Recipe{
		Id:       Id(id),
		Content:  content,
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
