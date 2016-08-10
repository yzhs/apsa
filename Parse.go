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
	"strings"
)

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
			metadataTypes := []string{"Quelle", "Tags", "Portionen",
			"Zubereitungszeit", "Kochzeit", "Backzeit", "Wartezeit",
			"Gesamtzeit", "Umluft", "Ober- und Unterhitze"}
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

	return Recipe{Id: Id(id),
		Content: content,
		Title: title,
		Portions: data["Portionen"],
		Source: data["Quelle"],
		Tags: parseTags(data["Tags"]),

		CookingTime: data["Kochzeit"],
		BakingTime: data["Backzeit"],
		WaitingTime: data["Wartezeit"],
		TotalTime: data["Gesamtzeit"],
		PreparationTime: data["Zubereitungszeit"],

		FanTemp: data["Umfluft"],
		TopAndBottomHeatTemp: data["Ober- und Unterhitze"],
		// TODO ingredients
	}
}
