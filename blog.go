package main

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/gorilla/feeds"
)

type blog struct {
	Title         string
	BaseDirectory string
	BaseURL       string
	Config        *config

	Entries []*entry
	Groups  []*group
	Tags    []*tag

	templates *templates
}

func newBlog(cfg *config) *blog {
	return &blog{
		Title:         cfg.Title,
		BaseDirectory: cfg.BaseDirectory,
		BaseURL:       cfg.BaseURL,
		Config:        cfg,
		Entries:       []*entry{},
		Groups:        []*group{},
		Tags:          []*tag{},
	}
}

func (b *blog) regenerate() error {
	measure(b.readTemplates, fail, "read templates in %vms.")

	measure(b.readEntries, fail, "found entries in %vms.")
	measure(b.writeEntries, fail, "rendered %v entries in %vms.", len(b.Entries))

	measure(b.findGroups, fail, "found groups in %vms.")
	measure(b.renderGroups, fail, "rendered %v groups in %vms.", len(b.Groups))

	measure(b.findTags, fail, "found tags in %vms.")
	measure(b.renderTags, fail, "rendered %v tags in %vms.", len(b.Tags))

	measure(b.renderFeed, fail, "rendered feed in %vms.")

	measure(b.renderMainIndex, fail, "rendered main index in %vms.")

	measure(b.renderSitemap, fail, "rendered sitemap in %vms.")

	return nil
}

func (b *blog) readTemplates() error {
	var err error
	b.templates, err = readTemplates(b.Config.Templates)
	return err
}

func (b *blog) collectURLs() []string {
	urls := []string{}

	urls = append(urls, b.URL())

	for _, e := range b.Entries {
		urls = append(urls, e.URL())
	}

	for _, g := range b.Groups {
		urls = append(urls, g.URL())
	}

	for _, t := range b.Tags {
		urls = append(urls, t.URL())
	}

	return urls
}

func (b *blog) URL() string {
	return urlJoin(b.BaseURL, "index.html")
}

func (b *blog) renderSitemap() error {
	if b.Config.SitemapFile == "" {
		verbose("no sitemap file configured")
		return nil
	}

	var buf bytes.Buffer
	var err error

	xml := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
{{ range . }}    <url>
        <loc>{{ . }}</loc>
    </url>
{{ end }}</urlset>
`

	tmpl, err := template.New("sitemap").Parse(xml)
	if err != nil {
		return err
	}

	urls := b.collectURLs()
	err = tmpl.ExecuteTemplate(&buf, "sitemap", urls)
	if err != nil {
		return fmt.Errorf("failed to execute sitemap template: %w", err)
	}

	fn := filepath.Join(b.BaseDirectory, b.Config.SitemapFile)
	err = ioutil.WriteFile(fn, buf.Bytes(), 0755)
	if err != nil {
		return fmt.Errorf("failed to write %#v: %w", fn, err)
	}

	verbose("write sitemap to %#v with %v entries.", fn, len(urls))
	return nil
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

	for _, e := range b.LatestEntries(3) {
		itm := &feeds.Item{
			Title:   e.Title,
			Link:    &feeds.Link{Href: e.URL()},
			Source:  &feeds.Link{Href: e.URL()},
			Created: e.Posted,
			Author:  &feeds.Author{Name: e.Author},
			Content: string(e.RenderedHTML),
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

func (b *blog) findTagNames() []string {
	uniq := map[string]struct{}{}

	for _, e := range b.Entries {
		for _, tn := range e.Tags {
			uniq[tn] = struct{}{}
		}
	}

	result := make([]string, 0, len(uniq))
	for tn, _ := range uniq {
		result = append(result, tn)
	}
	sort.Strings(result)

	return result
}

func (b *blog) findTags() error {
	tagNames := b.findTagNames()

	b.Tags = make([]*tag, 0, len(tagNames))
	for _, tn := range tagNames {
		t := newTag(b, tn)
		b.Tags = append(b.Tags, t)
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

func (b *blog) findGroupNames() []string {
	uniqNames := map[string]struct{}{}

	for _, e := range b.Entries {
		uniqNames[e.Group()] = struct{}{}
	}

	result := make([]string, 0, len(uniqNames))
	for n, _ := range uniqNames {
		result = append(result, n)
	}

	sort.Strings(result)

	return result
}

func (b *blog) findGroups() error {
	groupNames := b.findGroupNames()

	b.Groups = make([]*group, 0, len(groupNames))
	for _, n := range groupNames {
		g := newGroup(b, n)
		b.Groups = append(b.Groups, g)
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

func (b *blog) readEntries() error {
	mds := []string{}
	walker := func(pth string, info os.FileInfo, err error) error {
		if filepath.Ext(info.Name()) == ".md" {
			mds = append(mds, pth)
		}
		return err
	}
	err := filepath.Walk(b.BaseDirectory, walker)
	if err != nil {
		return fmt.Errorf("failed to search for md files: %w", err)
	}
	verbose("walked base-dir %#v and found %v md files.", b.BaseDirectory, len(mds))

	b.Entries = make([]*entry, 0, len(mds))
	for _, md := range mds {
		e, err := newEntry(b, md)
		if err != nil {
			return err
		}
		b.Entries = append(b.Entries, e)
	}

	sortByDate(b.Entries)
	return nil
}

func (b *blog) writeEntries() error {
	for _, e := range b.Entries {
		err := e.writeHTML()
		if err != nil {
			return err
		}
	}
	sortByDate(b.Entries)
	return nil
}

func (b *blog) LatestEntries(count int) []*entry {
	result := make([]*entry, 0, count)

	if len(b.Entries) == 0 {
		return result
	} else if len(b.Entries) < count {
		count = len(b.Entries)
	}

	for i := 0; i < count; i++ {
		result = append(result, b.Entries[i])
	}
	return result
}
