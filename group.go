package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"text/template"
)

type group struct {
	Name            string
	GroupDirectory  string
	RelativeLink    string
	RenderedEntries []*entry
	Entries         []*entry
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

	fp := filepath.Join(g.GroupDirectory, "index.html")
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
