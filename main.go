package main

import (
	"flag"
	"fmt"
	"os"
)

// TODO: rss

func main() {
	var err error
	args := newArguments(os.Args[1:])

	err = args.parse()
	fail(err)

	err = args.validate()
	fail(err)

	lg := newLog(args.Title, args.BaseDirectory)
	err = lg.regenerate()
	fail(err)
}

type arguments struct {
	raw []string

	Title         string `cli:"title"`
	BaseDirectory string `cli:"base-dir"`
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

	return flags.Parse(a.raw)
}
