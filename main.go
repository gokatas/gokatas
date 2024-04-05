package main

import (
	"encoding/json"
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

const (
	katasFile = "gokatas.json"
	reposURL  = "https://api.github.com/orgs/gokatas/repos"
)

func main() {
	log.SetPrefix(os.Args[0] + ": ")
	log.SetFlags(0)

	stats := flag.Bool("stats", false, "show your statistics")
	done := flag.String("done", "", "you've just done `kata`")
	flag.Parse()

	if *stats {
		ss, err := getStats(katasFile)
		if err != nil {
			log.Fatal(err)
		}
		printStats(ss)
		os.Exit(0)
	}

	katas, err := getKatas(reposURL)
	if err != nil {
		log.Fatal(err)
	}
	katas = filter(katas, show)

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

		if err := storeStats(*done, katasFile); err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}

	sort.Sort(byGoLines(katas))
	print(katas)
}

type Stats map[string][]time.Time

func printStats(ss Stats) {
	const format = "%v\t%v\t%v\n"
	tw := new(tabwriter.Writer).Init(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintf(tw, format, "Kata", "Done", "Last")
	fmt.Fprintf(tw, format, "----", "----", "----")
	for kata, dones := range ss {
		fmt.Fprintf(tw, format, kata, fmt.Sprintf("%dx", len(dones)), dones[len(dones)-1].Format("2006-01-02 15:04:05"))
	}
	tw.Flush()
}

func storeStats(kata, file string) error {
	fi, err := os.Stat(file)
	if err != nil {
		return err
	}
	var stats Stats
	var in []byte
	if fi.Size() > 0 {
		in, err = os.ReadFile(file)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(in, &stats); err != nil {
			return err
		}
	} else {
		stats = make(map[string][]time.Time)
	}
	stats[kata] = append(stats[kata], time.Now())
	data, err := json.Marshal(stats)
	if err != nil {
		return err
	}
	return os.WriteFile(file, data, 0644)
}

func getStats(file string) (Stats, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var ss Stats
	if err := json.Unmarshal(data, &ss); err != nil {
		return nil, err
	}
	return ss, nil
}

func cloneAndCount(k kata) (lines int, err error) {
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

func filter(katas []kata, fn func(kata) bool) []kata {
	var filtered []kata
	for _, k := range katas {
		if fn(k) {
			filtered = append(filtered, k)
		}
	}
	return filtered
}

func show(k kata) bool {
	var hidden = []string{".github", "gokatas"}
	for _, name := range hidden {
		if k.Name == name {
			return false
		}
	}
	return true
}

type kata struct {
	Name        string   `json:"name"`
	SshUrl      string   `json:"ssh_url"`
	HtmlUrl     string   `json:"html_url"`
	CloneUrl    string   `json:"clone_url"`
	Stars       int      `json:"stargazers_count"`
	Topics      []string `json:"topics"` // standard library packages
	Description string   `json:"description"`
	goLines     int
}

type byGoLines []kata

func (x byGoLines) Len() int { return len(x) }
func (x byGoLines) Less(i, j int) bool {
	if x[i].goLines == x[j].goLines {
		return x[i].Name < x[j].Name
	}
	return x[i].goLines < x[j].goLines
}
func (x byGoLines) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

func print(katas []kata) {
	const format = "%v\t%v\t%v\t%v\t%v\n"
	tw := new(tabwriter.Writer).Init(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintf(tw, format, "Lines", "Name", "Description", "Standard library packages", "URL")
	fmt.Fprintf(tw, format, "-----", "----", "-----------", "-------------------------", "---")
	for _, k := range katas {
		fmt.Fprintf(tw, format,
			k.goLines,
			k.Name,
			k.Description,
			strings.Join(k.Topics, " "),
			k.CloneUrl)
	}
	tw.Flush()
}

func getKatas(url string) ([]kata, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var katas []kata
	if err := json.Unmarshal(b, &katas); err != nil {
		return nil, err
	}
	return katas, nil
}
