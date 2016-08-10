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
	"html/template"
	"os"
)

const (
	NAME    = "Apsa"
	VERSION = "0.1"
)

// Statistics concerning the size of the library.
type Stats struct {
	num_recipes int
	file_size   int64
}

func (s Stats) Num() int {
	return s.num_recipes
}

func (s Stats) Size() int64 {
	return s.file_size
}

type Id string

type Backend interface {
	Search(query []string) ([]Id, error)
	GenerateIndex() error
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
	CacheDirectory     string
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
