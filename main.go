package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	cal := flag.Bool("cal", false, "print also activity calendar")
	done := flag.String("done", "", "write down `kata` you've done")
	doneFile := flag.String("donefile", filepath.Join(home, "gokatas.json"), "where to write down katas you've done")
	explain := flag.String("explain", "", "use AI to explain `kata`")
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

	if *cal {
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
