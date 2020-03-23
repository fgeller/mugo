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

type blog struct {
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

func newBlog(cfg *config) *blog {
	return &blog{
		Title:           cfg.Title,
		BaseDirectory:   cfg.BaseDirectory,
		BaseURL:         cfg.BaseURL,
		Config:          cfg,
		Entries:         []*entry{},
		RenderedEntries: []*entry{},
		Groups:          map[string]*group{},
		Tags:            map[string]*tag{},
	}
}

func (b *blog) regenerate() error {
	measure(b.readTemplates, fail, "read templates in %vms.")

	measure(b.findEntries, fail, "found entries in %vms.")
	measure(b.renderEntries, fail, "rendered %v entries in %vms.", len(b.Entries))

	measure(b.findGroups, fail, "found groups in %vms.")
	measure(b.renderGroups, fail, "rendered %v groups in %vms.", len(b.Groups))

	measure(b.findTags, fail, "found tags in %vms.")
	measure(b.renderTags, fail, "rendered %v tags in %vms.", len(b.Tags))

	measure(b.renderFeed, fail, "rendered feed in %vms.")

	measure(b.renderMainIndex, fail, "rendered main index in %vms.")

	return nil
}

func (b *blog) readTemplates() error {
	var err error
	b.templates, err = readTemplates(b.Config.Templates)
	return err
}

func (b *blog) renderFeed() error {
	fc := b.Config.Feed
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

	for i, e := range b.RenderedEntries {
		if i >= 3 {
			break
		}

		url := b.BaseURL
		if "/" != url[len(url)-1:] {
			url += "/"
		}
		url += e.MainToEntryPath()

		itm := &feeds.Item{
			Title:   e.Title,
			Link:    &feeds.Link{Href: url},
			Source:  &feeds.Link{Href: url},
			Created: e.Date,
			Author:  &feeds.Author{Name: e.Author},
			Content: e.RenderedHTML,
		}
		fd.Add(itm)
	}

	if fc.RSSEnabled {
		of := filepath.Join(b.BaseDirectory, "rss.xml")
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
		of := filepath.Join(b.BaseDirectory, "atom.xml")
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

func (b *blog) findTags() error {
	for _, e := range b.Entries {
		for _, t := range e.Tags {
			_, ok := b.Tags[t]
			if ok {
				b.Tags[t].Entries = append(b.Tags[t].Entries, e)
			} else {
				td := &tag{
					Name:            t,
					RelativeLink:    filepath.Base(b.BaseDirectory),
					TagDirectory:    b.BaseDirectory,
					Entries:         []*entry{e},
					RenderedEntries: []*entry{},
					Blog:            b,
					template:        b.templates.Tags,
				}
				b.Tags[t] = td
			}
			if !e.IsDraft {
				b.Tags[t].RenderedEntries = append(b.Tags[t].RenderedEntries, e)
			}
		}
	}

	for _, t := range b.Tags {
		sortByDate(t.Entries)
		sortByDate(t.RenderedEntries)
	}

	return nil
}

func (b *blog) renderTags() error {
	for _, t := range b.Tags {
		err := t.renderIndex()
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *blog) findGroups() error {
	for _, e := range b.Entries {
		gp := e.groupPath()
		gn := filepath.Base(gp)

		pth := filepath.Join(filepath.Base(b.BaseDirectory), gn)
		fp := filepath.Join(b.BaseDirectory, gn)

		_, ok := b.Groups[gn]
		if !ok {
			g := &group{
				Name:            gn,
				GroupDirectory:  fp,
				RelativeLink:    pth,
				Entries:         []*entry{e},
				RenderedEntries: []*entry{},
				Blog:            b,
				template:        b.templates.Group,
			}
			b.Groups[gn] = g
		} else {
			b.Groups[gn].Entries = append(b.Groups[gn].Entries, e)
		}

		if !e.IsDraft {
			b.Groups[gn].RenderedEntries = append(b.Groups[gn].RenderedEntries, e)
		}

	}

	for _, g := range b.Groups {
		sortByDate(g.Entries)
		sortByDate(g.RenderedEntries)
	}

	return nil
}

func (b *blog) renderGroups() error {
	for _, g := range b.Groups {
		err := g.renderIndex()
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *blog) renderMainIndex() error {
	var err error
	var buf bytes.Buffer

	err = b.templates.Main.ExecuteTemplate(&buf, "main", b)
	if err != nil {
		return fmt.Errorf("failed to execute main index template: %w", err)
	}

	fp := filepath.Join(b.BaseDirectory, "index.html")
	err = ioutil.WriteFile(fp, buf.Bytes(), 0777)
	if err != nil {
		return fmt.Errorf("failed to write main index file: %w", err)
	}
	verbose("rendered main index.")

	return nil
}

func (b *blog) LatestRenderedEntry() *entry {
	if len(b.RenderedEntries) == 0 {
		return nil
	} else {
		return b.RenderedEntries[0]
	}
}

func (b *blog) findEntries() error {
	walker := func(pth string, info os.FileInfo, err error) error {
		if filepath.Ext(info.Name()) == ".md" {
			e := &entry{
				MDFile:   pth,
				HTMLFile: htmlPath(pth),
				Blog:     b,
				template: b.templates.Entry,
			}
			b.Entries = append(b.Entries, e)
		}
		return err
	}
	err := filepath.Walk(b.BaseDirectory, walker)
	verbose("walked base-dir %#v and found %v entries.", b.BaseDirectory, len(b.Entries))
	return err
}

func (b *blog) renderEntries() error {
	for _, e := range b.Entries {
		err := e.render()
		if err != nil {
			return err
		}
		if !e.IsDraft {
			b.RenderedEntries = append(b.RenderedEntries, e)
		}
	}
	sortByDate(b.Entries)
	sortByDate(b.RenderedEntries)
	return nil
}
