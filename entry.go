package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"

	"github.com/fgeller/relabs"
)

type entry struct {
	MDFile   string
	HTMLFile string

	Title    string
	Summary  template.HTML
	Posted   time.Time
	Modified time.Time
	Author   string
	Tags     []string

	RenderedHTML template.HTML

	Blog *blog
}

func newEntry(b *blog, md string) (*entry, error) {
	e := &entry{MDFile: md, Blog: b}

	html, err := inferHTMLFilePath(b, md)
	if err != nil {
		return nil, err
	}
	e.HTMLFile = html

	err = e.readModified()
	if err != nil {
		return nil, err
	}

	return e, e.readMD()
}

func (e *entry) readModified() error {
	mf, err := os.Open(e.MDFile)
	if err != nil {
		return err
	}

	st, err := mf.Stat()
	if err != nil {
		return err
	}

	e.Modified = st.ModTime()

	return nil
}

func (e *entry) parseHeader(ctx parser.Context) error {
	header := meta.Get(ctx)
	var err error
	var ok bool

	e.Title, ok = header["title"].(string)
	if !ok {
		return fmt.Errorf("title is missing in %#v", e.MDFile)
	}

	e.Author, ok = header["author"].(string)
	if !ok {
		return fmt.Errorf("author is missing in %#v", e.MDFile)
	}

	e.Posted, err = time.Parse("2006-01-02", header["date"].(string))
	if err != nil {
		return fmt.Errorf("failed to parse header date in %#v: %w", e.MDFile, err)
	}

	e.Tags = []string{}
	raw, ok := header["tags"].([]interface{})
	if !ok {
		return fmt.Errorf("tags are not passed as array of strings in %#v: %w", e.MDFile, err)
	}
	for _, t := range raw {
		e.Tags = append(e.Tags, t.(string))
	}

	_, ok = header["summary"].(string)
	if ok {
		e.Summary = template.HTML(header["summary"].(string))
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

func (e *entry) RelativeURL() string {
	return urlJoin("/", e.Group(), e.Dir(), e.HTMLFileName())
}

func (e *entry) BaseURL() (*url.URL, error) {
	raw := urlJoin(e.Blog.BaseURL, e.Group(), e.Dir()) + "/"
	return url.Parse(raw)
}

func (e *entry) readMD() error {
	exts := []goldmark.Extender{meta.Meta, extension.GFM}
	if e.Blog.Config.ResolveRelativeLinks {
		bu, err := e.BaseURL()
		if err != nil {
			return err
		}
		exts = append(exts, relabs.NewRelabs(bu))
	}

	md := goldmark.New(
		goldmark.WithExtensions(exts...),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)
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

	if e.Summary == "" {
		reader := text.NewReader(src)
		doc := md.Parser().Parse(reader)

		var p bytes.Buffer
		err := md.Renderer().Render(&p, src, doc.FirstChild())
		if err == nil {
			e.Summary = template.HTML(p.String())
		}
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

	err = ioutil.WriteFile(e.HTMLFile, buf.Bytes(), 0644)
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
