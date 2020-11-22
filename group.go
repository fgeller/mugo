package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"
)

type group struct {
	Name     string
	Entries  []*entry
	Blog     *blog
	Modified time.Time
}

func newGroup(b *blog, name string) *group {
	g := &group{Name: name, Entries: []*entry{}, Blog: b}
	for _, e := range b.Entries {
		if e.Group() == name {
			g.Entries = append(g.Entries, e)
		}
	}
	sortByDate(g.Entries)
	g.Modified = findLatestModified(g.Entries)

	return g
}

func (g *group) URL() string {
	return urlJoin(g.Blog.BaseURL, g.Name, g.HTMLFileName())
}

func (g *group) RelativeURL() string {
	return urlJoin("/", g.Name, g.HTMLFileName())
}

func (g *group) HTMLFileName() string {
	return "index.html"
}

func (g *group) renderIndex() error {
	var err error
	var buf bytes.Buffer

	err = g.Blog.templates.Group.ExecuteTemplate(&buf, "group", g)
	if err != nil {
		return fmt.Errorf("failed to execute group index template: %w", err)
	}

	fp := filepath.Join(g.Blog.OutputDirectory, g.Name, g.HTMLFileName())
	err = ioutil.WriteFile(fp, buf.Bytes(), 0777)
	if err != nil {
		return fmt.Errorf("failed to write group index file: %w", err)
	}
	verbose("rendered index for group %#v to %#v.", g.Name, fp)

	return nil
}
