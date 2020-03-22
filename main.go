package main

import (
	"flag"
	"fmt"
	"os"
)

// TODO: rss
// TODO: template/html
// TODO: fsnotify regenerate dir

func main() {
	var err error
	var tpls *templates
	args := newArguments(os.Args[1:])

	err = args.parse()
	fail(err)

	err = args.validate()
	fail(err)

	tpls, err = readTemplates(args)
	fail(err)

	lg := newLog(args.Title, args.BaseDirectory, tpls)
	err = lg.regenerate()
	fail(err)
}

type arguments struct {
	raw []string

	Title         string `cli:"title"`
	BaseDirectory string `cli:"base-dir"`

	MainTemplate  string `cli:"main-template"`
	GroupTemplate string `cli:"group-template"`
	TagsTemplate  string `cli:"tags-template"`
	EntryTemplate string `cli:"entry-template"`
}

func newArguments(raw []string) *arguments {
	return &arguments{raw: raw}
}

func (a *arguments) validate() error {
	if a.Title == "" {
		return fmt.Errorf("title is required")
	}
	if a.BaseDirectory == "" {
		return fmt.Errorf("base-dir is required")
	}
	return nil
}

func (a *arguments) parse() error {
	flags := flag.NewFlagSet("mugo", flag.ContinueOnError)

	flags.StringVar(&a.Title, "title", "", "Top-level title (required).")
	flags.StringVar(&a.BaseDirectory, "base-dir", "", "Base directory to scan (required).")

	flags.StringVar(&a.MainTemplate, "main-template", "", "Go template for the main index page.")
	flags.StringVar(&a.GroupTemplate, "group-template", "", "Go template for the group index page.")
	flags.StringVar(&a.TagsTemplate, "tags-template", "", "Go template for the tags index page.")
	flags.StringVar(&a.EntryTemplate, "entry-template", "", "Go template for the entry page.")

	return flags.Parse(a.raw)
}
