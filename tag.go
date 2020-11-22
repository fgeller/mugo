package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type tag struct {
	Name     string
	Entries  []*entry
	Blog     *blog
	Modified time.Time
}

func newTag(b *blog, name string) *tag {
	t := &tag{Name: name, Entries: []*entry{}, Blog: b}
	for _, e := range b.Entries {
		for _, tn := range e.Tags {
			if tn == name {
				t.Entries = append(t.Entries, e)
				break
			}
		}
	}
	sortByDate(t.Entries)
	t.Modified = findLatestModified(t.Entries)

	return t
}

func (t *tag) URL() string {
	return urlJoin(t.Blog.BaseURL, "tags", t.HTMLFileName())
}

func (t *tag) RelativeURL() string {
	return urlJoin("/", "tags", t.HTMLFileName())
}

func (t *tag) HTMLFileName() string {
	return fmt.Sprintf("%s.html", t.Name)
}

func (t *tag) renderIndex() error {
	var err error
	var buf bytes.Buffer

	err = t.Blog.templates.Tags.ExecuteTemplate(&buf, "tags", t)
	if err != nil {
		return fmt.Errorf("failed to execute tag index template: %w", err)
	}

	tagDir := filepath.Join(t.Blog.OutputDirectory, "tags")
	err = os.Mkdir(tagDir, 0770)
	if !os.IsExist(err) {
		return fmt.Errorf("failed to create tags directory [%s] err=%w", tagDir, err)
	}

	fp := filepath.Join(tagDir, t.HTMLFileName())
	err = ioutil.WriteFile(fp, buf.Bytes(), 0777)
	if err != nil {
		return fmt.Errorf("failed to write tag index file: %w", err)
	}
	verbose("rendered index for tag %#v to %#v.", t.Name, fp)

	return nil
}
