package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"text/tabwriter"
)

func main() {
	log.SetPrefix(os.Args[0] + ": ")
	log.SetFlags(0)

	katas, err := list()
	if err != nil {
		log.Fatal(err)
	}
	print(katas)
}

type kata struct {
	Name    string   `json:"name"`
	Size    int      `json:"size"`
	SshUrl  string   `json:"ssh_url"`
	HtmlUrl string   `json:"html_url"`
	Stars   int      `json:"stargazers_count"`
	Topics  []string `json:"topics"`
}

func print(katas []kata) {
	const format = "%v\t%v\t%v\t%v\t%v\n"
	tw := new(tabwriter.Writer).Init(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintf(tw, format, "Name", "Size", "Stars", "Topics", "URL")
	fmt.Fprintf(tw, format, "----", "----", "-----", "------", "---")
	for _, k := range katas {
		fmt.Fprintf(tw, format, k.Name, k.Size, k.Stars, k.Topics, k.HtmlUrl)
	}
	tw.Flush()
}

func list() ([]kata, error) {
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
