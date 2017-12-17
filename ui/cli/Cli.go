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

func printStats() {
	stats := ComputeStatistics()
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

	switch {
	case index:
		err := GenerateIndex()
		TryLogError(err)
	case stats:
		printStats()
	case version:
		fmt.Println(NAME, VERSION)
	default:
		i := 1
		if len(os.Args) > 0 {
			if os.Args[1] == "--" {
				i += 1
			} else if os.Args[1] == "all" {
				fmt.Printf("Rendered all %d recipes.\n", RenderAllRecipes())
				os.Exit(0)
			}
		}
		cmd := exec.Command("firefox", "http://localhost/apsa/search?q="+
			strings.Join(os.Args[i:], " "))
		TryLogError(cmd.Run())
	}
}
