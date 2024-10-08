package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const reposURL = "https://api.github.com/orgs/gokatas/repos"

var doneFile string

func main() {
	log.SetPrefix(os.Args[0] + ": ")
	log.SetFlags(0)

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	done := flag.String("done", "", "write down `kata` you've done")
	doneFile := flag.String("donefile", filepath.Join(home, "gokatas.json"), "where to write down katas you've done")
	explain := flag.String("explain", "", "use AI to explain `kata`")
	report := flag.Bool("report", false, "print also activity report")
	sortby := flag.String("sortby", "name", "sort by `column`")
	wide := flag.Bool("wide", false, "print wider output")
	flag.Parse()

	katas, err := getKatas(reposURL)
	if err != nil {
		log.Fatal(err)
	}
	katas = filter(katas, show)

	if err := katas.getStats(*doneFile); err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	for i := range katas {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			lines, err := cloneAndCount(katas[i])
			if err != nil {
				log.Fatal(err)
			}
			katas[i].goLines = lines
		}(i)
	}
	wg.Wait()

	if *explain != "" {
		if err := katas.explain(*explain); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	if *done != "" {
		var found bool
		for _, k := range katas {
			if k.Name == *done {
				found = true
				if time.Since(lastTime(k.done)).Seconds() < 300 {
					fmt.Print("Last done less than 5 mins ago. Are you sure? ")
					var input string
					fmt.Scanln(&input)
					if !strings.HasPrefix(strings.ToLower(input), "y") {
						os.Exit(1)
					}
				}
				break
			}
		}
		if !found {
			log.Fatalf("no such kata: %s", *done)
		}
		if err := katas.storeStats(*doneFile, *done); err != nil {
			log.Fatal(err)
		}
	}

	sortKatas(katas, sortby)
	printKatas(katas, wide)

	if *report {
		fmt.Println()
		until := time.Now()
		since := time.Now().Add(-time.Hour * 24 * 90)
		if *wide {
			since = time.Now().Add(-time.Hour * 24 * 180)
		}
		b := Boundary{
			Since: since,
			Until: until,
		}
		printReport(katas, b)
	}
}
