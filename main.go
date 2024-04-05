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

var statsFile string

func main() {
	log.SetPrefix(os.Args[0] + ": ")
	log.SetFlags(0)

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	statsFile = filepath.Join(home, "gokatas.json")
	flag.StringVar(&statsFile, "statsfile", statsFile, "where to keep stats")

	done := flag.String("done", "", "you've just done `kata`")
	stats := flag.Bool("stats", false, "show what you've done")
	flag.Parse()

	if *stats {
		ss, err := getStats(statsFile)
		if err != nil {
			log.Fatal(err)
		}
		ss.Print()
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

		ss, err := getStats(statsFile)
		if err != nil {
			log.Fatal(err)
		}
		if err := ss.Store(*done); err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}

	sort.Sort(byGoLines(katas))
	print(katas)
}

type Stats struct {
	File string
	Done map[string][]time.Time
}

func getStats(file string) (*Stats, error) {
	stats := Stats{
		File: file,
		Done: make(map[string][]time.Time),
	}
	data, err := os.ReadFile(file)
	if err != nil {
		switch {
		case errors.Is(err, fs.ErrNotExist):
			return &stats, nil
		default:
			return nil, err
		}
	}
	if err := json.Unmarshal(data, &stats.Done); err != nil {
		return nil, err
	}
	return &stats, nil
}

func (stats *Stats) Print() {
	const format = "%v\t%v\t%v\n"
	tw := new(tabwriter.Writer).Init(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintf(tw, format, "Kata", "Done", "Last done")
	fmt.Fprintf(tw, format, "----", "----", "---------")
	for kata, dones := range stats.Done {
		fmt.Fprintf(tw, format, kata, fmt.Sprintf("%dx", len(dones)), dones[len(dones)-1].Format("2006-01-02 15:04:05"))
	}
	tw.Flush()
}

func (stats *Stats) Store(kata string) error {
	_, err := os.Stat(stats.File)
	if err != nil {
		switch {
		case errors.Is(err, fs.ErrNotExist):
			f, err := os.Create(statsFile)
			if err != nil {
				return err
			}
			f.Close()
		default:
			return err
		}
	} else {
		data, err := os.ReadFile(stats.File)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, &stats.Done); err != nil {
			return err
		}
	}

	stats.Done[kata] = append(stats.Done[kata], time.Now())
	data, err := json.Marshal(stats.Done)
	if err != nil {
		return err
	}
	return os.WriteFile(stats.File, data, 0644)
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
