package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/feeds"
)

type log struct {
	Title         string
	BaseDirectory string
	BaseURL       string
	Config        *config

	Entries         []*entry
	RenderedEntries []*entry
	Groups          map[string]*group
	Tags            map[string]*tag

	templates *templates
}

func newLog(cfg *config) *log {
	lg := &log{
		Title:           cfg.Title,
		BaseDirectory:   cfg.BaseDirectory,
		BaseURL:         cfg.BaseURL,
		Config:          cfg,
		Entries:         []*entry{},
		RenderedEntries: []*entry{},
		Groups:          map[string]*group{},
		Tags:            map[string]*tag{},
	}
	return lg
}

func (l *log) regenerate() error {
	measure(l.readTemplates, fail, "read templates in %vms.")

	measure(l.findEntries, fail, "found entries in %vms.")
	measure(l.renderEntries, fail, "rendered %v entries in %vms.", len(l.Entries))

	measure(l.findGroups, fail, "found groups in %vms.")
	measure(l.renderGroups, fail, "rendered %v groups in %vms.", len(l.Groups))

	measure(l.findTags, fail, "found tags in %vms.")
	measure(l.renderTags, fail, "rendered %v tags in %vms.", len(l.Tags))

	measure(l.renderFeed, fail, "rendered feed in %vms.")

	measure(l.renderMainIndex, fail, "rendered main index in %vms.")

	return nil
}

func (l *log) readTemplates() error {
	var err error
	l.templates, err = readTemplates(l.Config.Templates)
	return err
}

func (l *log) renderFeed() error {
	fc := l.Config.Feed
	if fc == nil {
		verbose("no config for rendering feed.")
		return nil
	}

	fd := &feeds.Feed{
		Title:       fc.Title,
		Link:        &feeds.Link{Href: fc.LinkHREF},
		Description: fc.Description,
		Author:      &feeds.Author{Name: fc.AuthorName, Email: fc.AuthorEmail},
		Created:     time.Now(),
	}

	for i, e := range l.RenderedEntries {
		if i >= 3 {
			break
		}

		url := l.BaseURL
		if "/" != url[len(url)-1:] {
			url += "/"
		}
		url += e.MainToEntryPath()

		itm := &feeds.Item{
			Title:   e.Title,
			Link:    &feeds.Link{Href: fmt.Sprintf("%s/%s", l.BaseURL, e.MainToEntryPath())},
			Source:  &feeds.Link{Href: fmt.Sprintf("%s/%s", l.BaseURL, e.MainToEntryPath())},
			Created: e.Date,
			Author:  &feeds.Author{Name: e.Author},
			Content: e.RenderedHTML,
		}
		fd.Add(itm)
	}

	if fc.RSSEnabled {
		of := filepath.Join(l.BaseDirectory, "rss.xml")
		fh, err := os.OpenFile(of, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755)
		if err != nil {
			return fmt.Errorf("failed to open rss file %#v: %w", of, err)
		}

		w := bufio.NewWriter(fh)
		err = fd.WriteRss(w)
		if err != nil {
			return fmt.Errorf("failed to write feed to rss: %w", err)
		}
	}

	if fc.AtomEnabled {
		of := filepath.Join(l.BaseDirectory, "atom.xml")
		fh, err := os.OpenFile(of, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755)
		if err != nil {
			return fmt.Errorf("failed to open atom file %#v: %w", of, err)
		}

		w := bufio.NewWriter(fh)
		err = fd.WriteAtom(w)
		if err != nil {
			return fmt.Errorf("failed to write feed to atom: %w", err)
		}
	}

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
					RelativeLink:    filepath.Base(l.BaseDirectory),
					TagDirectory:    l.BaseDirectory,
					Entries:         []*entry{e},
					RenderedEntries: []*entry{},
					Log:             l,
					template:        l.templates.Tags,
				}
				l.Tags[t] = td
			}
			if !e.IsDraft {
				l.Tags[t].RenderedEntries = append(l.Tags[t].RenderedEntries, e)
			}
		}
	}

	for _, t := range l.Tags {
		sortByDate(t.Entries)
		sortByDate(t.RenderedEntries)
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
		fp := filepath.Join(l.BaseDirectory, gn)

		_, ok := l.Groups[gn]
		if !ok {
			g := &group{
				Name:            gn,
				GroupDirectory:  fp,
				RelativeLink:    pth,
				Entries:         []*entry{e},
				RenderedEntries: []*entry{},
				Log:             l,
				template:        l.templates.Group,
			}
			l.Groups[gn] = g
		} else {
			l.Groups[gn].Entries = append(l.Groups[gn].Entries, e)
		}

		if !e.IsDraft {
			l.Groups[gn].RenderedEntries = append(l.Groups[gn].RenderedEntries, e)
		}

	}

	for _, g := range l.Groups {
		sortByDate(g.Entries)
		sortByDate(g.RenderedEntries)
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

	err = l.templates.Main.ExecuteTemplate(&buf, "main", l)
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

func (l *log) LatestRenderedEntry() *entry {
	if len(l.RenderedEntries) == 0 {
		return nil
	} else {
		return l.RenderedEntries[0]
	}
}

func (l *log) findEntries() error {
	walker := func(pth string, info os.FileInfo, err error) error {
		if filepath.Ext(info.Name()) == ".md" {
			e := &entry{
				MDFile:   pth,
				HTMLFile: htmlPath(pth),
				Log:      l,
				template: l.templates.Entry,
			}
			l.Entries = append(l.Entries, e)
		}
		return err
	}
	err := filepath.Walk(l.BaseDirectory, walker)
	verbose("walked base-dir %#v and found %v entries.", l.BaseDirectory, len(l.Entries))
	return err
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
	sortByDate(l.Entries)
	sortByDate(l.RenderedEntries)
	return nil
}
