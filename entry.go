package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"path/filepath"
	"sort"
	"time"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
)

type entry struct {
	MDFile   string
	HTMLFile string

	Title   string
	Summary string
	Posted  time.Time
	Author  string
	Tags    []string

	RenderedHTML template.HTML

	Blog *blog
}

func newEntry(b *blog, md string) (*entry, error) {
	e := &entry{
		MDFile:   md,
		HTMLFile: htmlPath(md),
		Blog:     b,
	}

	return e, e.readMD()
}

func (e *entry) parseHeader(ctx parser.Context) error {
	header := meta.Get(ctx)
	var err error

	e.Title = header["title"].(string)

	e.Author = header["author"].(string)

	e.Posted, err = time.Parse("2006-01-02", header["date"].(string))
	if err != nil {
		return fmt.Errorf("failed to parse header date %w", err)
	}

	e.Tags = []string{}
	raw, ok := header["tags"].([]interface{})
	if !ok {
		return fmt.Errorf("tags are not passed as array of strings: %w", err)
	}
	for _, t := range raw {
		e.Tags = append(e.Tags, t.(string))
	}

	_, ok = header["summary"].(string)
	if ok {
		e.Summary = header["summary"].(string)
	}

	return nil
}

func (e *entry) Group() string {
	gp := filepath.Dir(filepath.Dir(e.MDFile))
	return filepath.Base(gp)
}

func (e *entry) Dir() string {
	return filepath.Base(filepath.Dir(e.MDFile))
}

func (e *entry) HTMLFileName() string {
	return filepath.Base(e.HTMLFile)
}

func (e *entry) URL() string {
	return urlJoin(e.Blog.BaseURL, e.Group(), e.Dir(), e.HTMLFileName())
}

func (e *entry) readMD() error {
	md := goldmark.New(goldmark.WithExtensions(meta.Meta))
	ctx := parser.NewContext()
	var buf bytes.Buffer

	src, err := ioutil.ReadFile(e.MDFile)
	if err != nil {
		return err
	}

	err = md.Convert(src, &buf, parser.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to convert markdown to html: %w", err)
	}
	e.RenderedHTML = template.HTML(buf.String())

	err = e.parseHeader(ctx)
	if err != nil {
		return fmt.Errorf("failed to parse header: %w", err)
	}
	return nil
}

func (e *entry) writeHTML() error {
	var err error
	var buf bytes.Buffer

	err = e.Blog.templates.Entry.ExecuteTemplate(&buf, "entry", e)
	if err != nil {
		return fmt.Errorf("failed to execute entry template: %w", err)
	}

	err = ioutil.WriteFile(e.HTMLFile, buf.Bytes(), 0777)
	if err != nil {
		return err
	}
	verbose("write entry %#v to %#v.", e.Title, e.HTMLFile)

	return nil
}

func sortByDate(entries []*entry) {
	chrono := func(i, j int) bool { return entries[i].Posted.After(entries[j].Posted) }
	sort.Slice(entries, chrono)
}
