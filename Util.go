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
	"fmt"
	"io/ioutil"
	"os"
)

func LogError(err interface{}) {
	fmt.Fprintln(os.Stderr, err)
}

// TryLogError logs an error if `err != nil`
func TryLogError(err interface{}) {
	if err != nil {
		LogError(err)
	}
}

// Load the content of a given recipe from disk.
func readRecipe(id Id) (string, error) {
	result, err := ioutil.ReadFile(Config.KnowledgeDirectory + string(id) + ".md")
	return string(result), err
}

// Load the content of a template file with the given name.
func readTemplate(filename string) (string, error) {
	result, err := ioutil.ReadFile(Config.TemplateDirectory + filename + ".html")
	return string(result), err
}

// Write a HTML file with the given name and content to Apsa's temp
// directory.
func writeTemp(id Id, data string) error {
	return ioutil.WriteFile(Config.TempDirectory+string(id)+".html", []byte(data), 0644)
}

// Compute the combined size of all files in a given directory.
func getDirSize(dir string) (int, int64) {
	directory, err := os.Open(dir)
	TryLogError(err)
	defer directory.Close()
	fileInfo, err := directory.Readdir(0)
	if err != nil {
		panic(err)
	}
	result := int64(0)
	for _, file := range fileInfo {
		result += file.Size()
	}
	return len(fileInfo), result
}

// Get the time a given file was last modified as a Unix time.
func getModTime(file string) (int64, error) {
	info, err := os.Stat(file)
	if err != nil {
		return -1, err
	}
	return info.ModTime().Unix(), nil
}

// Cache the newest modification of any of the template files as a Unix time
// (i.e. seconds since 1970-01-01).
var templatesModTime int64 = -1

// All recognized template files
// TODO Generate the list⁈
var templateFiles = []string{"header.html", "footer.html"}

// Check whether a given recipe has to be recompiled
func isUpToDate(id Id) bool {
	if templatesModTime == -1 {
		// Check template for modification times
		templatesModTime = 0

		for _, file := range templateFiles {
			foo, err := getModTime(Config.TemplateDirectory + file)
			if err != nil {
				break
			}
			if foo > templatesModTime {
				templatesModTime = foo
			}
		}
	}

	info, err := os.Stat(Config.CacheDirectory + string(id) + ".html")
	if err != nil {
		return false
	}
	imageTime := info.ModTime().Unix()

	if imageTime < templatesModTime {
		return false
	}

	info, err = os.Stat(Config.KnowledgeDirectory + string(id) + ".md")
	if err != nil {
		return false // When in doubt, recompile
	}
	recipeTime := info.ModTime().Unix()

	return imageTime > recipeTime
}
