package main

import (
	"encoding/json"
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
)

func main() {
	log.SetPrefix(os.Args[0] + ": ")
	log.SetFlags(0)

	katas, err := get()
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

	sort.Sort(byGoLines(katas))
	print(katas)
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

func (x byGoLines) Len() int           { return len(x) }
func (x byGoLines) Less(i, j int) bool { return x[i].goLines < x[j].goLines }
func (x byGoLines) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

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

func get() ([]kata, error) {
	url := "https://api.github.com/orgs/gokatas/repos"
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
