package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/fgeller/relabs"
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

type top struct {
	MDFile   string
	HTMLFile string

	Title    string
	Modified time.Time

	RenderedHTML template.HTML

	Blog *blog
}

func newTop(b *blog, md string) (*top, error) {
	t := &top{MDFile: md, Blog: b}

	html, err := inferHTMLFilePath(b, md)
	if err != nil {
		return nil, err
	}
	t.HTMLFile = html

	err = t.readModified()
	if err != nil {
		return nil, err
	}

	return t, t.readMD()
}

func (t *top) readModified() error {
	mf, err := os.Open(t.MDFile)
	if err != nil {
		return err
	}

	st, err := mf.Stat()
	if err != nil {
		return err
	}

	t.Modified = st.ModTime()

	return nil
}

func (t *top) BaseURL() (*url.URL, error) {
	raw := urlJoin(t.Blog.BaseURL, t.Dir()) + "/"
	return url.Parse(raw)
}

func (t *top) Dir() string {
	return filepath.Base(filepath.Dir(t.MDFile))
}

func (t *top) readMD() error {
	exts := []goldmark.Extender{meta.Meta, extension.GFM}
	if t.Blog.Config.ResolveRelativeLinks {
		bu, err := t.BaseURL()
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

	src, err := os.ReadFile(t.MDFile)
	if err != nil {
		return err
	}

	err = md.Convert(src, &buf, parser.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to convert markdown to html: %w", err)
	}
	t.RenderedHTML = template.HTML(buf.String())

	err = t.parseHeader(ctx)
	if err != nil {
		return fmt.Errorf("failed to parse header: %w", err)
	}

	return nil
}

func (t *top) parseHeader(ctx parser.Context) error {
	header := meta.Get(ctx)
	var ok bool

	t.Title, ok = header["title"].(string)
	if !ok {
		return fmt.Errorf("title is missing in %#v", t.MDFile)
	}

	return nil
}

func (t *top) writeHTML() error {
	var err error
	var buf bytes.Buffer

	err = t.Blog.templates.Top.ExecuteTemplate(&buf, "top", t)
	if err != nil {
		return fmt.Errorf("failed to execute top template: %w", err)
	}

	err = os.WriteFile(t.HTMLFile, buf.Bytes(), 0644)
	if err != nil {
		return err
	}
	verbose("write top %#v to %#v.", t.Title, t.HTMLFile)

	return nil
}
