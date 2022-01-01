// Apsa
//
// Copyright (C) 2015,2016  Colin Benner
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

package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime/pprof"
	"strings"

	flag "github.com/ogier/pflag"

	backend "github.com/yzhs/apsa"
)

const SOCKET_PATH = "/tmp/apsa.sock"

// Generate a HTML file describing the size of the library.
func printStats() string {
	stats := backend.ComputeStatistics()
	n := stats.Num()
	size := float32(stats.Size()) / 1024.0
	return fmt.Sprintf("The library contains %v recipes with a total size of %.1f kiB.\n", n, size)
}

// Send the statistics page to the client.
func statsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, printStats())
}

// Handle the edit-link, causing the browser to open that recipe in an editor.
func editHandler(w http.ResponseWriter, r *http.Request) {
	headers := w.Header()
	headers["Content-Type"] = []string{"application/x-backend-edit"}
	id := r.FormValue("id")
	fmt.Fprintf(w, id)
}

// Serve the search page.
func mainHandler(w http.ResponseWriter, r *http.Request) {
	html, err := loadHTMLTemplate("main")
	if err != nil {
		fmt.Fprintf(w, "%v", err)
		return
	}
	fmt.Fprintln(w, string(html))
}

func loadHTMLTemplate(name string) ([]byte, error) {
	return ioutil.ReadFile(backend.Config.TemplateDirectory + name + ".html")
}

type match struct {
	backend.Recipe
	SourceURL template.HTML
}

type result struct {
	Query        string
	Matches      []backend.Recipe
	NumMatches   int
	TotalMatches int
}

func renderTemplate(w io.Writer, templateName string, resultData result) {
	funcMap := template.FuncMap{
		"link": func(x string) template.HTML {
			if strings.HasPrefix(x, "http://") || strings.HasPrefix(x, "https://") {
				return template.HTML("<a href=\"" + x + "\">" + x + "</a>")
			}
			return template.HTML(template.HTMLEscapeString(x))
		},
	}
	tmplFile := backend.Config.TemplateDirectory + templateName + ".html"
	t, err := template.New(templateName + ".html").Funcs(funcMap).ParseFiles(tmplFile)
	if err != nil {
		fmt.Fprintf(w, "Error: %v", err)
		return
	}
	err = t.ExecuteTemplate(w, templateName+".html", resultData)
	backend.TryLogError(err)
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

// Load the rendered content of a given recipe from disk.
func readRecipe(id backend.Id) (string, error) {
	result, err := ioutil.ReadFile(backend.Config.CacheDirectory + string(id) + ".html")
	return string(result), err
}

type Controller struct {
	searchEngine backend.SearchEngine
}

// Handle a query and serve the results.
func (c Controller) queryHandler(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("q")
	if query == "" {
		mainHandler(w, r)
		return
	}

	results, err := c.searchEngine.Search(query)
	if err != nil {
		panic(err)
	}
	numMatches := len(results.Ids)

	for i := range results.Ids {
		tmp, err := readRecipe(results.Ids[i].Id)
		if err != nil {
			panic(err)
		}
		results.Ids[i].HTML = template.HTML(tmp)
	}
	data := result{
		Query: query, NumMatches: numMatches, Matches: results.Ids[:min(20, numMatches)],
		TotalMatches: results.Total,
	}
	renderTemplate(w, "search", data)
}

func serveDirectory(prefix string, directory string) {
	http.Handle(prefix, http.StripPrefix(prefix, http.FileServer(http.Dir(directory))))
}

func main() {
	var profile, version bool
	flag.BoolVarP(&version, "version", "v", false, "\tShow version")
	flag.BoolVar(&profile, "profile", false, "\tEnable profiler")
	flag.Parse()

	backend.InitConfig()
	backend.Config.MaxResults = 20

	if profile {
		f, err := os.Create("backend.prof")
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// TODO run BuildIndex() when there is something new

	if version {
		fmt.Println(backend.NAME, backend.VERSION)
		return
	}

	controller := Controller{backend.CreateSearchEngine()}

	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/stats", statsHandler)
	http.HandleFunc("/search", controller.queryHandler)
	http.HandleFunc("/backend.apsaedit", editHandler)
	serveDirectory("/images/", backend.Config.CacheDirectory)
	serveDirectory("/static/", backend.Config.TemplateDirectory+"static")
	server := http.Server{}

	listener, err := net.Listen("unix", SOCKET_PATH)
	if err != nil {
		backend.LogError(err)
		return
	}
	defer listener.Close()
	os.Chmod(SOCKET_PATH, 0777)

	err = server.Serve(listener)
	backend.TryLogError(err)
	os.Remove(SOCKET_PATH)
}
