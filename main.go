package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	cfg, err := readConfig()
	fail(err)

	lg := newBlog(cfg)
	err = lg.regenerate()
	fail(err)
}

type config struct {
	Title           string   `json:"title"`
	BaseDirectory   string   `json:"base-directory"`
	OutputDirectory string   `json:"output-directory"`
	OutputExcludes  []string `json:"output-excludes"`
	BaseURL         string   `json:"base-url"`

	SitemapFile string `json:"sitemap-file"`

	ResolveRelativeLinks bool `json:"resolve-relative-links"`

	Templates *templatesConfig `json:"templates"`
	Feed      *feedConfig      `json:"feed"`
}

type templatesConfig struct {
	Main  string `json:"main"`
	Group string `json:"group"`
	Tags  string `json:"tags"`
	Entry string `json:"entry"`
}

type feedConfig struct {
	RSSEnabled  bool   `json:"rss-enabled"`
	AtomEnabled bool   `json:"atom-enabled"`
	Title       string `json:"title"`
	LinkHREF    string `json:"link-href"`
	AuthorName  string `json:"author-name"`
	AuthorEmail string `json:"author-email"`
	Description string `json:"description"`
}

func (c *config) validate() error {
	if c.Title == "" {
		return fmt.Errorf("title is required")
	}
	if c.BaseDirectory == "" {
		return fmt.Errorf("base-directory is required")
	}
	if c.BaseURL == "" {
		return fmt.Errorf("base-url is required")
	}
	return nil
}

func readConfig() (*config, error) {
	cf, err := readFlags()
	if err != nil {
		return nil, err
	}

	bt, err := ioutil.ReadFile(cf)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %#v: %w", cf, err)
	}

	result := &config{}
	err = json.Unmarshal(bt, result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file %#v: %w", cf, err)
	}

	err = result.validate()
	if err != nil {
		return nil, err
	}

	log.Printf("read config: %+v\n", result)

	return result, nil
}

func readFlags() (string, error) {
	var err error
	var cf string
	flags := flag.NewFlagSet("mugo", flag.ContinueOnError)
	flags.StringVar(&cf, "config", "", "Path to JSON config file (required).")

	err = flags.Parse(os.Args[1:])
	if err != nil {
		return "", err
	}

	if cf == "" {
		return "", fmt.Errorf("config is required.")
	}

	return cf, nil
}
