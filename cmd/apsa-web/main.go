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
	"github.com/russross/blackfriday"

	backend "github.com/yzhs/apsa"
)

const SocketPath = "/tmp/apsa.sock"

// Send the statistics page to the client.
func (c Controller) statsHandler(w http.ResponseWriter, r *http.Request) {
	stats := c.searchEngine.ComputeStatistics()
	n := stats.Num()
	size := float32(stats.Size()) / 1024.0
	fmt.Fprintf(w, "The library contains %v recipes with a total size of %.1f kiB.\n", n, size)
}

// Handle the edit-link, causing the browser to open that recipe in an editor.
func editHandler(w http.ResponseWriter, r *http.Request) {
	headers := w.Header()
	headers["Content-Type"] = []string{"application/x-apsa-edit"}
	id := r.FormValue("id")
	fmt.Fprint(w, id)
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

type Result struct {
	Query        string
	Matches      []backend.ModernistRecipe
	NumMatches   int
	TotalMatches int
}

var funcMap = template.FuncMap{
	"link": func(x string) template.HTML {
		if strings.HasPrefix(x, "http://") || strings.HasPrefix(x, "https://") {
			return template.HTML("<a href=\"" + x + "\">" + x + "</a>")
		}
		return template.HTML(template.HTMLEscapeString(x))
	},
	"markdown": func(x string) template.HTML {
		r := blackfriday.MarkdownCommon([]byte(x))
		return template.HTML(r)
	},
}

func renderTemplate(w io.Writer, templateName string, resultData Result) {
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
	numMatches := len(results.Recipes)

	data := Result{
		Query: query, NumMatches: numMatches, Matches: results.Recipes[:min(20, numMatches)],
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
		f, err := os.Create("apsa.prof")
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if version {
		fmt.Println(backend.NAME, backend.VERSION)
		return
	}

	controller := Controller{backend.NewSearchEngine()}

	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/stats", controller.statsHandler)
	http.HandleFunc("/search", controller.queryHandler)
	http.HandleFunc("/apsa.apsaedit", editHandler)
	serveDirectory("/static/", backend.Config.TemplateDirectory+"static")
	server := http.Server{}

	listener, err := net.Listen("unix", SocketPath)
	if err != nil {
		backend.LogError(err)
		return
	}
	defer listener.Close()
	os.Chmod(SocketPath, 0777)

	err = server.Serve(listener)
	backend.TryLogError(err)
	os.Remove(SocketPath)
}
