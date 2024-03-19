package main

import (
	"encoding/json"
	"flag"
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

	s := flag.Int("s", 1, "sort by column")
	flag.Parse()

	katas, err := get()
	if err != nil {
		log.Fatal(err)
	}
	katas = filter(katas, show)
	order(katas, *s)
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

type customSort struct {
	katas []kata
	less  func(x, y kata) bool
}

func (x customSort) Len() int           { return len(x.katas) }
func (x customSort) Less(i, j int) bool { return x.less(x.katas[i], x.katas[j]) }
func (x customSort) Swap(i, j int)      { x.katas[i], x.katas[j] = x.katas[j], x.katas[i] }

func order(katas []kata, column int) {
	sort.Sort(customSort{katas, func(x, y kata) bool {
		switch column {
		case 1:
			if x.Name != y.Name {
				return x.Name < y.Name
			}
		case 2:
			if x.Size != y.Size {
				return x.Size < y.Size
			}
		case 3:
			if x.Stars != y.Stars {
				return x.Stars > y.Stars
			}
		default:
			log.Fatal("select column 1, 2 or 3")
		}

		// secondary sort
		if x.Name != y.Name {
			return x.Name < y.Name
		}

		return false
	}})
}

func print(katas []kata) {
	const format = "%v\t%v\t%v\t%v\t%v\t%v\n"
	tw := new(tabwriter.Writer).Init(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintf(tw, format, "Name", "Size", "Stars", "Description", "Topics", "URL")
	fmt.Fprintf(tw, format, "----", "----", "-----", "-----------", "------", "---")
	for _, k := range katas {
		fmt.Fprintf(tw, format,
			k.Name,
			k.Size,
			k.Stars,
			k.Description,
			strings.Join(k.Topics, ", "),
			k.HtmlUrl)
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
