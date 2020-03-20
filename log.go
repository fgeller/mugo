package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

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

func (l *log) regenerate() error {
	measure(l.findEntries, fail, "found entries in %vms.")
	measure(l.renderEntries, fail, "rendered %v entries in %vms.", len(l.Entries))

	measure(l.findGroups, fail, "found groups in %vms.")
	measure(l.renderGroups, fail, "rendered %v groups in %vms.", len(l.Groups))

	measure(l.findTags, fail, "found tags in %vms.")
	measure(l.renderTags, fail, "rendered %v tags in %vms.", len(l.Tags))

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