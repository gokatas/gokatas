package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/sashabaranov/go-openai"
)

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

type Katas []Kata

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
		case "desc", "description":
			if x.Description != y.Description {
				return x.Description < y.Description
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
			log.Fatalf("we don't sort by %s here", *column)
		}
		if x.Name != y.Name {
			return x.Name < y.Name
		}
		return false
	}})
}

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
		return fmt.Errorf("parsing %s: %v", file, err)
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

	for i := range katas {
		if katas[i].Name == kata {
			katas[i].done = append(katas[i].done, time.Now())
		}
	}

	stats := make(map[string][]time.Time)
	var data []byte
	for _, kata := range katas {
		stats[kata.Name] = kata.done
		data, err = json.MarshalIndent(stats, "", "  ")
		if err != nil {
			return err
		}
	}
	if err := os.WriteFile(donefile, data, 0644); err != nil {
		return err
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

func printKatas(katas Katas, wide *bool) {
	tw := new(tabwriter.Writer).Init(os.Stdout, 0, 8, 2, ' ', 0)
	if *wide {
		const format = "%v\t%v\t%v\t%v\t%v\t%v\t%v\n"
		fmt.Fprintf(tw, format, "Name", "Description", "Lines", "Done", "Last done", "URL", "Standard library packages")
		fmt.Fprintf(tw, format, "----", "-----------", "-----", "----", "---------", "---", "-------------------------")
		for _, k := range katas {
			fmt.Fprintf(tw, format,
				k.Name,
				k.Description,
				k.goLines,
				fmt.Sprintf("%dx", len(k.done)),
				humanize(lastTime(k.done)),
				k.CloneUrl,
				strings.Join(k.Topics, " "),
			)
		}

	} else {
		const format = "%v\t%v\t%v\t%v\t%v\n"
		fmt.Fprintf(tw, format, "Name", "Description", "Lines", "Done", "Last done")
		fmt.Fprintf(tw, format, "----", "-----------", "-----", "----", "---------")
		for _, k := range katas {
			fmt.Fprintf(tw, format,
				k.Name,
				k.Description,
				k.goLines,
				fmt.Sprintf("%dx", len(k.done)),
				humanize(lastTime(k.done)),
			)
		}
	}
	tw.Flush()
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

func (katas Katas) explain(name string) error {
	token := os.Getenv("OPENAI_API_KEY")
	if token == "" {
		return fmt.Errorf("set OPENAI_API_KEY environment variable")
	}
	client := openai.NewClient(token)

	var kata Kata
	var found bool
	for _, k := range katas {
		if name == k.Name {
			kata = k
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("no such kata: %s", name)
	}

	input, err := getKataContent(kata.CloneUrl)
	if err != nil {
		return err
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt + input,
				},
			},
		},
	)

	if err != nil {
		return err
	}

	fmt.Println(resp.Choices[0].Message.Content)
	return nil
}

func getKataContent(kataUrl string) (string, error) {
	path, err := clone(kataUrl)
	if err != nil {
		return "", fmt.Errorf("cloning %s: %v", path, err)
	}

	code, err := getGoCode(path)
	if err != nil {
		return "", err
	}
	return code, nil
}

func getGoCode(path string) (string, error) {
	var content string
	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && filepath.Ext(path) == ".go" {
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			content += string(b)
		}
		return nil
	})
	return content, err
}

func clone(kataUrl string) (path string, err error) {
	u, err := url.Parse(kataUrl)
	if err != nil {
		return "", err
	}

	path, err = os.MkdirTemp("", "kata")
	if err != nil {
		return path, err
	}

	err = exec.Command("git", "clone", u.String(), path).Run()
	if err != nil {
		return path, fmt.Errorf("executing 'git clone %s %s': %v", kataUrl, path, err)
	}

	return path, nil
}
