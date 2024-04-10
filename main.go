package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"
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
	doneFile = filepath.Join(home, "gokatas.json")
	flag.StringVar(&doneFile, "donefile", doneFile, "where to keep katas you've done")
	done := flag.String("done", "", "you've just done `kata`")
	sortby := flag.String("sortby", "name", "sort by `column`")
	wide := flag.Bool("wide", false, "show all columns")
	flag.Parse()

	katas, err := getKatas(reposURL)
	if err != nil {
		log.Fatal(err)
	}
	katas = filter(katas, show)

	if err := katas.getStats(doneFile); err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	for i := range katas {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lines, err := cloneAndCount(katas[i])
			if err != nil {
				log.Fatal(err)
			}
			katas[i].goLines = lines
		}()
	}
	wg.Wait()

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
		if err := katas.storeStats(doneFile, *done); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	sortKatas(katas, sortby)
	printKatas(katas, wide)
}

type Katas []Kata

func (katas Katas) getStats(file string) error {
	data, err := os.ReadFile(file)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	done := make(map[string][]time.Time)
	if err := json.Unmarshal(data, &done); err != nil {
		return err
	}
	for i := range katas {
		katas[i].done = done[katas[i].Name]
	}
	return nil
}

func (katas Katas) storeStats(donefile, kata string) error {
	_, err := os.Stat(donefile)
	if err != nil {
		switch {
		case errors.Is(err, fs.ErrNotExist):
			f, err := os.Create(doneFile)
			if err != nil {
				return err
			}
			f.Close()
		default:
			return err
		}
	}

	for _, k := range katas {
		if k.Name == kata {
			k.done = append(k.done, time.Now())
			stats := make(map[string][]time.Time)
			stats[kata] = k.done
			data, err := json.Marshal(stats)
			if err != nil {
				return err
			}
			if err := os.WriteFile(donefile, data, 0644); err != nil {
				return err
			}
		}
	}

	return nil
}

func cloneAndCount(k Kata) (lines int, err error) {
	dir, err := os.MkdirTemp("", "kata")
	if err != nil {
		return
	}
	defer os.RemoveAll(dir)

	err = exec.Command("git", "clone", k.CloneUrl, dir).Run()
	if err != nil {
		return
	}

	return countGo(dir)
}

func countGo(dir string) (lines int, err error) {
	visit := func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && filepath.Ext(path) == ".go" {
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			n := len(strings.Split(string(b), "\n"))
			lines += n
		}
		return nil
	}
	err = filepath.WalkDir(dir, visit)
	return
}

func filter(katas Katas, fn func(Kata) bool) []Kata {
	var filtered []Kata
	for _, k := range katas {
		if fn(k) {
			filtered = append(filtered, k)
		}
	}
	return filtered
}

func show(k Kata) bool {
	var hidden = []string{".github", "gokatas"}
	for _, name := range hidden {
		if k.Name == name {
			return false
		}
	}
	return true
}

type Kata struct {
	Name        string   `json:"name"`
	SshUrl      string   `json:"ssh_url"`
	HtmlUrl     string   `json:"html_url"`
	CloneUrl    string   `json:"clone_url"`
	Stars       int      `json:"stargazers_count"`
	Topics      []string `json:"topics"` // standard library packages
	Description string   `json:"description"`
	goLines     int
	done        []time.Time
}

func printKatas(katas Katas, wide *bool) {
	tw := new(tabwriter.Writer).Init(os.Stdout, 0, 8, 2, ' ', 0)
	if *wide {
		const format = "%v\t%v\t%v\t%v\t%v\t%v\t%v\n"
		fmt.Fprintf(tw, format, "Name", "Lines", "Done", "Last done", "URL", "Description", "Standard library packages")
		fmt.Fprintf(tw, format, "----", "-----", "----", "---------", "---", "-----------", "-------------------------")
		for _, k := range katas {
			fmt.Fprintf(tw, format,
				k.Name,
				k.goLines,
				fmt.Sprintf("%dx", len(k.done)),
				humanize(lastTime(k.done)),
				k.CloneUrl,
				k.Description,
				strings.Join(k.Topics, " "),
			)
		}

	} else {
		const format = "%v\t%v\t%v\t%v\t%v\n"
		fmt.Fprintf(tw, format, "Name", "Lines", "Done", "Last done", "URL")
		fmt.Fprintf(tw, format, "----", "-----", "----", "---------", "---")
		for _, k := range katas {
			fmt.Fprintf(tw, format,
				k.Name,
				k.goLines,
				fmt.Sprintf("%dx", len(k.done)),
				humanize(lastTime(k.done)),
				k.CloneUrl,
			)
		}
	}
	tw.Flush()
}

func lastTime(times []time.Time) time.Time {
	var last time.Time
	for _, t := range times {
		if t.After(last) {
			last = t
		}
	}
	return last
}

// humanize makes the time easier to read for humans.
func humanize(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	daysAgo := int(time.Since(t).Hours() / 24)
	w := "day"
	if daysAgo != 1 {
		w += "s"
	}
	return fmt.Sprintf("%d %s ago", daysAgo, w)
}

func getKatas(url string) (Katas, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var katas Katas
	if err := json.Unmarshal(b, &katas); err != nil {
		return nil, err
	}
	return katas, nil
}

type customSort struct {
	katas Katas
	less  func(x, y Kata) bool
}

func (x customSort) Len() int           { return len(x.katas) }
func (x customSort) Less(i, j int) bool { return x.less(x.katas[i], x.katas[j]) }
func (x customSort) Swap(i, j int)      { x.katas[i], x.katas[j] = x.katas[j], x.katas[i] }

// sortKatas sorts katas by column. Not all columns are sortable. Secondary sort
// orders is always by kata name.
func sortKatas(katas Katas, column *string) {
	sort.Sort(customSort{katas, func(x, y Kata) bool {
		switch strings.ToLower(*column) {
		case "name":
			if x.Name != y.Name {
				return x.Name < y.Name
			}
		case "lines":
			if x.goLines != y.goLines {
				return x.goLines < y.goLines
			}
		case "done":
			if len(x.done) != len(y.done) {
				return len(x.done) < len(y.done)
			}
		case "last", "last done":
			if lastTime(x.done) != lastTime(y.done) {
				return lastTime(x.done).After(lastTime(y.done))
			}
		default:
			log.Fatalf("why would you sort by %s", *column)
		}
		if x.Name != y.Name {
			return x.Name < y.Name
		}
		return false
	}})
}
