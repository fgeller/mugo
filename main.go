package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"text/template"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
)

func main() {
	var err error

	lg := newLog(
		"felix/log",
		"/Users/fgeller/src/github.com/fgeller/web/log/",
	)

	err = lg.regenerate()
	if err != nil {
		panic(err)
	}
}

//  log
//   |
//   > index.html
//   > 2020
//      |
//      > index.html
//      2020-02-25
//       |
//       > index.html (redirect)
//       > mist.html
//       > mist.jpg
//
//      2020-03-17
//       |
//       > index.html (redirect)
//       > emacs-27.html

type group struct {
	Name            string
	Path            string
	RenderedEntries []*entry
	Entries         []*entry
}

type tag struct {
	Name            string
	Path            string
	RenderedEntries []*entry
	Entries         []*entry
}

type log struct {
	Title         string
	BaseDirectory string

	Entries         []*entry
	RenderedEntries []*entry
	Groups          map[string]*group
	Tags            map[string]*tag
}

func newLog(title string, baseDir string) *log {
	lg := &log{
		Title:           title,
		BaseDirectory:   baseDir,
		Entries:         []*entry{},
		RenderedEntries: []*entry{},
		Groups:          map[string]*group{},
		Tags:            map[string]*tag{},
	}
	return lg
}

func measure(f func() error, eh func(error), mf string, args ...interface{}) {
	start := time.Now()
	err := f()
	if err != nil {
		eh(err)
	}
	elapsed := time.Since(start)
	args = append(args, elapsed.Milliseconds())
	verbose(mf, args...)
}

func fail(err error) {
	panic(err)
}

func (l *log) regenerate() error {
	measure(l.findEntries, fail, "found entries in %vms.")
	measure(l.renderEntries, fail, "rendered %v entries in %vms.", len(l.Entries))

	measure(l.findGroups, fail, "found groups in %vms.")
	measure(l.renderGroups, fail, "rendered %v groups in %vms.", len(l.Groups))

	measure(l.findTags, fail, "found tags in %vms.")
	measure(l.renderTags, fail, "rendered %v tags in %vms.", len(l.Tags))
	// TODO rss

	measure(l.renderMainIndex, fail, "rendered main index in %vms.")
	return nil
}

func (l *log) findTags() error {
	for _, e := range l.Entries {
		for _, t := range e.Tags {
			_, ok := l.Tags[t]
			if ok {
				l.Tags[t].Entries = append(l.Tags[t].Entries, e)
			} else {
				td := &tag{
					Name:            t,
					Path:            filepath.Base(l.BaseDirectory),
					Entries:         []*entry{e},
					RenderedEntries: []*entry{},
				}
				l.Tags[t] = td
			}
			if !e.IsDraft {
				l.Tags[t].RenderedEntries = append(l.Tags[t].RenderedEntries, e)
			}
		}
	}
	return nil
}

func (l *log) renderTags() error {
	for _, t := range l.Tags {
		err := t.renderIndex()
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *log) findGroups() error {
	for _, e := range l.Entries {
		gp := e.groupPath()
		gn := filepath.Base(gp)

		pth := filepath.Join(filepath.Base(l.BaseDirectory), gn)

		_, ok := l.Groups[gn]
		if !ok {
			g := &group{
				Name:            gn,
				Path:            pth,
				Entries:         []*entry{e},
				RenderedEntries: []*entry{},
			}
			l.Groups[gn] = g
		} else {
			l.Groups[gn].Entries = append(l.Groups[gn].Entries, e)
		}

		if !e.IsDraft {
			l.Groups[gn].RenderedEntries = append(l.Groups[gn].RenderedEntries, e)
		}

	}
	return nil
}

func (l *log) renderGroups() error {
	for _, g := range l.Groups {
		err := g.renderIndex()
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *log) renderMainIndex() error {
	var err error
	var buf bytes.Buffer

	t, err := template.New("main-index").Parse(tmplMain)
	if err != nil {
		return fmt.Errorf("failed to parse main index template: %w", err)
	}

	err = t.ExecuteTemplate(&buf, "main-index", l)
	if err != nil {
		return fmt.Errorf("failed to execute main index template: %w", err)
	}

	fp := filepath.Join(l.BaseDirectory, "index.html")
	err = ioutil.WriteFile(fp, buf.Bytes(), 0777)
	if err != nil {
		return fmt.Errorf("failed to write main index file: %w", err)
	}
	verbose("rendered main index.")

	return nil
}

func (t *tag) renderIndex() error {
	var err error
	var buf bytes.Buffer

	tmpl, err := template.New("tag-index").Parse(tmplTags)
	if err != nil {
		return fmt.Errorf("failed to parse tag index template: %w", err)
	}

	err = tmpl.ExecuteTemplate(&buf, "tag-index", t)
	if err != nil {
		return fmt.Errorf("failed to execute tag index template: %w", err)
	}

	fp := filepath.Join(t.Path, fmt.Sprintf("%s.html", t.Name))
	err = ioutil.WriteFile(fp, buf.Bytes(), 0777)
	if err != nil {
		return fmt.Errorf("failed to write tag index file: %w", err)
	}
	verbose("rendered index for tag %#v to %#v.", t.Name, fp)

	return nil
}

func (g *group) renderIndex() error {
	var err error
	var buf bytes.Buffer

	t, err := template.New("group-index").Parse(tmplGroup)
	if err != nil {
		return fmt.Errorf("failed to parse group index template: %w", err)
	}

	err = t.ExecuteTemplate(&buf, "group-index", g)
	if err != nil {
		return fmt.Errorf("failed to execute group index template: %w", err)
	}

	fp := filepath.Join(g.Path, "index.html")
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

func (l *log) findEntries() error {
	walker := func(pth string, info os.FileInfo, err error) error {
		if filepath.Ext(info.Name()) == ".md" {
			e := entry{
				MDPath:   pth,
				HTMLPath: htmlPath(pth),
			}
			l.Entries = append(l.Entries, &e)
		}
		return err
	}
	return filepath.Walk(l.BaseDirectory, walker)
}

func (l *log) renderEntries() error {
	for _, e := range l.Entries {
		err := e.render()
		if err != nil {
			return err
		}
		if !e.IsDraft {
			l.RenderedEntries = append(l.RenderedEntries, e)
		}
	}
	return nil
}

type entry struct {
	MDPath   string
	HTMLPath string

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
	return filepath.Dir(filepath.Dir(e.MDPath))
}

func (e *entry) Group() string {
	return filepath.Base(e.groupPath())
}

func (e *entry) GroupToEntryPath() string {
	date := filepath.Base(filepath.Dir(e.HTMLPath))
	fn := filepath.Base(e.HTMLPath)
	return filepath.Join(date, fn)
}

func (e *entry) MainToEntryPath() string {
	group := filepath.Base(filepath.Dir(filepath.Dir(e.HTMLPath)))
	date := filepath.Base(filepath.Dir(e.HTMLPath))
	fn := filepath.Base(e.HTMLPath)
	return filepath.Join(group, date, fn)
}

func (e *entry) render() error {
	md := goldmark.New(goldmark.WithExtensions(meta.Meta))
	ctx := parser.NewContext()
	var buf bytes.Buffer

	src, err := ioutil.ReadFile(e.MDPath)
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

	t, err := template.New("log-entry").Parse(templateEntry)
	if err != nil {
		return fmt.Errorf("failed to parse entry template: %w", err)
	}

	buf.Reset()
	err = t.ExecuteTemplate(&buf, "log-entry", e)
	if err != nil {
		return fmt.Errorf("failed to execute entry template: %w", err)
	}

	if !e.IsDraft {
		err = ioutil.WriteFile(e.HTMLPath, buf.Bytes(), 0777)
		if err != nil {
			return err
		}
	}

	return nil
}

func htmlPath(md string) string {
	bs := filepath.Base(md)
	fn := fmt.Sprintf("%s.html", bs[:len(bs)-len(".md")])
	html := filepath.Join(filepath.Dir(md), fn)
	return html
}

func verbose(fs string, args ...interface{}) {
	if true {
		fmt.Println(fmt.Sprintf(fs, args...))
	}
}

var tmplMain = `
<!doctype html>
<html>
<meta charset="UTF-8">
  <head>
    <title>log</title>
    <link rel="stylesheet" type="text/css" href="style.css">
  </head>

  <body>
    <header>
        <a href="index.html">log</a>
    </header>
    <section>
    Groups
    {{ range $gn, $g := .Groups }}
      <article>
        <a href="{{ $g.MainToGroupPath }}">{{ $gn }}</a>
        <br />
        {{ len $g.RenderedEntries }} entries
      </article>
    {{ end }}
    </section>
    <section>
    Tags
    {{ range $tn, $t := .Tags }}
      <article>
        <a href="{{ $tn }}.html">{{ $tn }}</a>
        <br />
        {{ len $t.RenderedEntries }} entries
      </article>
    {{ end }}
    </section>
  </body>
</html>
`

var tmplGroup = `
<!doctype html>
<html>
<meta charset="UTF-8">
  <head>
    <title>{{.Name}}</title>
    <link rel="stylesheet" type="text/css" href="../style.css">
  </head>

  <body>
    <header>
        <a href="../index.html">log</a> /
        {{ .Name }}
    </header>
    <section>
    {{ range .RenderedEntries }}
      <article>
        <a href="{{ .GroupToEntryPath }}">{{ .Title }}</a>
        <br />
        {{ range .Tags }}<a href="../{{ . }}.html">{{ . }}</a> {{ end}}
      </article>
    {{ end }}
    </section>
  </body>
</html>
`

var tmplTags = `
<!doctype html>
<html>
<meta charset="UTF-8">
  <head>
    <title>{{.Name}}</title>
    <link rel="stylesheet" type="text/css" href="style.css">
  </head>

  <body>
    <header>
        <a href="index.html">log</a> /
        {{ .Name }}
    </header>
    <section>
    {{ range .RenderedEntries }}
      <article>
        <a href="{{ .MainToEntryPath }}">{{ .Title }}</a>
        <br />
        {{ range .Tags }}<a href="{{ . }}.html">{{ . }}</a> {{ end}}
      </article>
    {{ end }}
    </section>
  </body>
</html>
`

var templateEntry = `
<!doctype html>
<html>
<meta charset="UTF-8">
  <head>
    <title>{{.Title}}</title>
    <link rel="stylesheet" type="text/css" href="../../style.css">
  </head>

  <body>
    <header>
        <a href="../../index.html">log</a> /
        <a href="../index.html">{{ .Group }}</a> /
        {{ .Title }}
    </header>

    <section>
      <article>
        {{.RenderedHTML}}
      </article>
    </section>
  </body>
</html>
`
