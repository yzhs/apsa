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
	// Search(query []string) ([]Id, error)
	//  ComputeStatistics() Statistics
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
	CacheDirectory     string
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
	Config.CacheDirectory = dir + "cache/"
	Config.TemplateDirectory = dir + "templates/"
	Config.TempDirectory = dir + "tmp/"
}

func CreateSearchEngine() SearchEngine {
	return Bleve{}
}
