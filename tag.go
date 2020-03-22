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
}

func (t *tag) renderIndex() error {
	var err error
	var buf bytes.Buffer

	tmpl, err := template.New("tag-index").Funcs(template.FuncMap{"FormatDate": FormatDate}).Parse(tmplTags)
	if err != nil {
		return fmt.Errorf("failed to parse tag index template: %w", err)
	}

	err = tmpl.ExecuteTemplate(&buf, "tag-index", t)
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
