package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"text/template"
	"time"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
)

type entry struct {
	MDFile   string
	HTMLFile string

	Title   string
	Date    time.Time
	Author  string
	Tags    []string
	IsDraft bool

	RenderedHTML string
}

func (e *entry) parseHeader(ctx parser.Context) error {
	header := meta.Get(ctx)
	var err error

	e.Title = header["title"].(string)

	e.Author = header["author"].(string)

	e.Date, err = time.Parse("2006-01-02", header["date"].(string))
	if err != nil {
		return fmt.Errorf("failed to parse header date %w", err)
	}

	e.Tags = []string{}
	raw := header["tags"].([]interface{})
	for _, t := range raw {
		e.Tags = append(e.Tags, t.(string))
	}

	_, ok := header["draft"].(bool)
	if ok {
		e.IsDraft = header["draft"].(bool)
	}

	return nil
}

func (e *entry) groupPath() string {
	return filepath.Dir(filepath.Dir(e.MDFile))
}

func (e *entry) Group() string {
	return filepath.Base(e.groupPath())
}

func (e *entry) GroupToEntryPath() string {
	date := filepath.Base(filepath.Dir(e.HTMLFile))
	fn := filepath.Base(e.HTMLFile)
	return filepath.Join(date, fn)
}

func (e *entry) MainToEntryPath() string {
	group := filepath.Base(filepath.Dir(filepath.Dir(e.HTMLFile)))
	date := filepath.Base(filepath.Dir(e.HTMLFile))
	fn := filepath.Base(e.HTMLFile)
	return filepath.Join(group, date, fn)
}

func (e *entry) render() error {
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
	e.RenderedHTML = buf.String()

	err = e.parseHeader(ctx)
	if err != nil {
		return fmt.Errorf("failed to parse header: %w", err)
	}

	t, err := template.New("log-entry").Parse(tmplEntry)
	if err != nil {
		return fmt.Errorf("failed to parse entry template: %w", err)
	}

	buf.Reset()
	err = t.ExecuteTemplate(&buf, "log-entry", e)
	if err != nil {
		return fmt.Errorf("failed to execute entry template: %w", err)
	}

	if !e.IsDraft {
		err = ioutil.WriteFile(e.HTMLFile, buf.Bytes(), 0777)
		if err != nil {
			return err
		}
		verbose("rendered entry %#v to %#v.", e.Title, e.HTMLFile)
	} else {
		verbose("rendered entry draft %#v in memory.", e.Title)
	}

	return nil
}
