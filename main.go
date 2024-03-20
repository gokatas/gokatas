package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
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
	sort.Sort(byName(katas))
	print(katas)
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
	Size        int      `json:"size"`
	SshUrl      string   `json:"ssh_url"`
	HtmlUrl     string   `json:"html_url"`
	Stars       int      `json:"stargazers_count"`
	Topics      []string `json:"topics"`
	Description string   `json:"description"`
}

type byName []kata

func (x byName) Len() int           { return len(x) }
func (x byName) Less(i, j int) bool { return x[i].Name < x[j].Name }
func (x byName) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

func print(katas []kata) {
	const format = "%v\t%v\t%v\t%v\n"
	tw := new(tabwriter.Writer).Init(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintf(tw, format, "Name", "Description", "Topics", "URL")
	fmt.Fprintf(tw, format, "----", "-----------", "------", "---")
	for _, k := range katas {
		fmt.Fprintf(tw, format,
			k.Name,
			k.Description,
			strings.Join(k.Topics, ", "),
			k.SshUrl)
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
