package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	flag "github.com/ogier/pflag"

	"github.com/yzhs/apsa"
)

func printStats(s apsa.SearchEngine) {
	stats := s.ComputeStatistics()
	n := stats.Num()
	size := float32(stats.Size()) / 1024.0
	fmt.Printf("The library contains %v recipes with a total size of %.1f kiB.\n", n, size)
}

func import_from_urls(urls []string) {
	for _, recipeUrl := range urls {
		if !strings.HasPrefix(recipeUrl, "http") {
			apsa.LogError(fmt.Sprintf("Could not import recipe from '%s'. Not recognized as an HTTP URL.", recipeUrl))
			continue
		}

		u, err := url.Parse(recipeUrl)
		if err != nil {
			apsa.LogError(fmt.Sprintf("Could not import recipe from '%s'. Not recognized as an HTTP URL.", recipeUrl))
			continue
		}

		hostname := strings.TrimPrefix(u.Hostname(), "www.")

		cmd := exec.Command("apsa-import-"+hostname, recipeUrl)
		apsa.TryLogError(cmd.Run())
	}
}

func main() {
	var index, profile, stats, version bool
	flag.BoolVarP(&index, "index", "i", false, "\tUpdate the index")
	flag.BoolVarP(&stats, "stats", "S", false, "\tPrint some statistics")
	flag.BoolVarP(&version, "version", "v", false, "\tShow version")
	flag.BoolVar(&profile, "profile", false, "\tEnable profiler")
	flag.Parse()

	if flag.Arg(0) == "import" {
		var args = flag.Args()[1:]
		import_from_urls(args)
		return
	}

	apsa.InitConfig()
	apsa.Config.MaxResults = 1e9

	searchEngine := apsa.NewSearchEngine()

	switch {
	case index:
		err := searchEngine.BuildIndex()
		apsa.TryLogError(err)
	case stats:
		printStats(searchEngine)
	case version:
		fmt.Println(apsa.NAME, apsa.VERSION)
	default:
		i := 1
		if len(os.Args) > 0 && os.Args[1] == "--" {
			i += 1
		}
		cmd := exec.Command(
			"firefox", "http://localhost/apsa/search?q="+
				strings.Join(os.Args[i:], " "),
		)
		apsa.TryLogError(cmd.Run())
	}
}
