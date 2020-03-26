package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"text/template"
)

type group struct {
	Name            string
	GroupDirectory  string
	RelativeLink    string
	RenderedEntries []*entry
	Entries         []*entry
	Blog            *blog
	template        *template.Template
}

func (g *group) URL() string {
	return urlJoin(g.Blog.BaseURL, g.Name, g.HTMLFileName())
}

func (g *group) HTMLFileName() string {
	return "index.html"
}

func (g *group) renderIndex() error {

	var err error
	var buf bytes.Buffer

	err = g.template.ExecuteTemplate(&buf, "group", g)
	if err != nil {
		return fmt.Errorf("failed to execute group index template: %w", err)
	}

	fp := filepath.Join(g.GroupDirectory, g.HTMLFileName())
	err = ioutil.WriteFile(fp, buf.Bytes(), 0777)
	if err != nil {
		return fmt.Errorf("failed to write group index file: %w", err)
	}
	verbose("rendered index for group %#v to %#v.", g.Name, fp)

	return nil
}

func (g *group) MainToGroupPath() string {
	return filepath.Join(g.Name, "index.html")
}
