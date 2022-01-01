package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime/pprof"
	"strings"

	flag "github.com/ogier/pflag"

	. "github.com/yzhs/apsa"
)

func printStats(s SearchEngine) {
	stats := s.ComputeStatistics()
	n := stats.Num()
	size := float32(stats.Size()) / 1024.0
	fmt.Printf("The library contains %v recipes with a total size of %.1f kiB.\n", n, size)
}

func main() {
	var index, profile, stats, version bool
	flag.BoolVarP(&index, "index", "i", false, "\tUpdate the index")
	flag.BoolVarP(&stats, "stats", "S", false, "\tPrint some statistics")
	flag.BoolVarP(&version, "version", "v", false, "\tShow version")
	flag.BoolVar(&profile, "profile", false, "\tEnable profiler")
	flag.Parse()

	InitConfig()
	Config.MaxResults = 1e9

	if profile {
		f, err := os.Create("apsa.prof")
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	searchEngine := CreateSearchEngine()

	switch {
	case index:
		err := searchEngine.BuildIndex()
		TryLogError(err)
	case stats:
		printStats(searchEngine)
	case version:
		fmt.Println(NAME, VERSION)
	default:
		i := 1
		if len(os.Args) > 0 && os.Args[1] == "--" {
			i += 1
		}
		cmd := exec.Command(
			"firefox", "http://localhost/apsa/search?q="+
				strings.Join(os.Args[i:], " "),
		)
		TryLogError(cmd.Run())
	}
}
