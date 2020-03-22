package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"text/template"
)

type tag struct {
	Name            string
	RelativeLink    string
	TagDirectory    string
	RenderedEntries []*entry
	Entries         []*entry
	template        *template.Template
}

func (t *tag) renderIndex() error {
	var err error
	var buf bytes.Buffer

	err = t.template.ExecuteTemplate(&buf, "tags", t)
	if err != nil {
		return fmt.Errorf("failed to execute tag index template: %w", err)
	}

	fp := filepath.Join(t.TagDirectory, fmt.Sprintf("%s.html", t.Name))
	err = ioutil.WriteFile(fp, buf.Bytes(), 0777)
	if err != nil {
		return fmt.Errorf("failed to write tag index file: %w", err)
	}
	verbose("rendered index for tag %#v to %#v.", t.Name, fp)

	return nil
}
