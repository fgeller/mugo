package main

import (
	"fmt"
	"io/ioutil"
	"text/template"
)

type templates struct {
	Main  *template.Template
	Group *template.Template
	Tags  *template.Template
	Entry *template.Template
}

func createTemplate(name, file, fallback string) (*template.Template, error) {
	var raw string
	funcs := template.FuncMap{"FormatDate": FormatDate}
	if file == "" {
		raw = fallback
		verbose("template %#v uses fallback rather than file: %#v", name, file)
	} else {
		byt, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read %v template: %w", name, err)
		}
		raw = string(byt)
		verbose("template %#v uses source from file %#v", name, file)
	}
	return template.New(name).Funcs(funcs).Parse(raw)
}

func readTemplates(cfg *templatesConfig) (*templates, error) {
	var err error
	result := &templates{}

	result.Main, err = createTemplate("main", cfg.Main, tmplMain)
	if err != nil {
		return nil, fmt.Errorf("failed to parse main template: %w", err)
	}

	result.Group, err = createTemplate("group", cfg.Group, tmplGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to parse group template: %w", err)
	}

	result.Tags, err = createTemplate("tags", cfg.Tags, tmplTags)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tags template: %w", err)
	}

	result.Entry, err = createTemplate("entry", cfg.Entry, tmplEntry)
	if err != nil {
		return nil, fmt.Errorf("failed to parse entry template: %w", err)
	}

	return result, nil
}

var tmplMain = `
<!doctype html>
<html>
<meta charset="UTF-8">
  <head>
    <title>{{ .Title }}</title>
    <link rel="stylesheet" type="text/css" href="style.css">
  </head>

  <body>
    <section class="main">
      <h1>{{ .Title }}</h1>
      <h2>latest entry</h2>
      <article>
      <div>
        <a href="{{ .LatestRenderedEntry.MainToEntryPath }}"><h2>{{ .LatestRenderedEntry.Title }}<h2></a>
      </div>
      <div>
        <div>
          tags: {{ range .LatestRenderedEntry.Tags }}<a href="{{ . }}.html">{{ . }}</a> {{ end}}
        </div>
        <div>
          posted on {{ FormatDate .LatestRenderedEntry.Date }}
        </div>
      </div>
      </article>
    </section>

    <section class="groups">
      <h2>tags</h2>
      {{ range $tn, $t := .Tags }}
        <article>
          <div>
            <a href="{{ $tn }}.html">{{ $tn }}</a> ({{ len $t.RenderedEntries }})
          </div>
          <div>
          </div>
        </article>
      {{ end }}

      <h2>groups</h2>
      {{ range $gn, $g := .Groups }}
        <article>
          <div>
            <a href="{{ $g.MainToGroupPath }}">{{ $gn }}</a> ({{ len $g.RenderedEntries }})
          </div>
          <div>
          </div>
        </article>
      {{ end }}

      <h2>feeds</h2>
        <article>
          <div>
            <a href="rss.xml">rss</a>
          </div>
          <div>
          </div>
        </article>

        <article>
          <div>
            <a href="atom.xml">atom</a>
          </div>
          <div>
          </div>
        </article>

      <h2>links</h2>
        <article>
          <div>
            github
          </div>
          <div>
          </div>
        </article>

        <article>
          <div>
            twitter
          </div>
          <div>
          </div>
        </article>

        <article>
          <div>
            insta
          </div>
          <div>
          </div>
        </article>

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

    <section>

      <h1>{{ .Name}}</h1>

      {{ range .RenderedEntries }}
      <article>
        <div>
          <a href="{{ .GroupToEntryPath }}"><h2>{{ .Title }}</h2></a>
        </div>
        <div>
          <div>
            posted on {{ FormatDate .Date }}
          </div>
          <div>
            tags: {{ range .Tags }}<a href="../{{ . }}.html">{{ . }}</a> {{ end}}
          </div>
          <div>
          </div>
        </div>
      </article>
      {{ end }}

    </section>

    <footer>
      <div>
        <a href="../index.html">{{ .Log.Title }}</a> /
        {{ .Name }}
      </div>
      <div>
        {{ len .RenderedEntries }} entries
      </div>
      <div>
      </div>
    </footer>

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

    <section>

      <h1>{{ .Name}}</h1>

      {{ range .RenderedEntries }}
      <article>
        <div>
          <a href="{{ .MainToEntryPath }}"><h2>{{ .Title }}</h2></a>
        </div>
        <div>
          <div>
            posted on {{ FormatDate .Date }}
          </div>
          <div>
            tags: {{ range .Tags }}<a href="{{ . }}.html">{{ . }}</a> {{ end}}
          </div>
          <div>
          </div>
        </div>
      </article>
      {{ end }}

    </section>

    <footer>
      <div>
        <a href="index.html">{{ .Log.Title }}</a> /
        {{ .Name }}
      </div>
      <div>
        {{ len .RenderedEntries }} entries
      </div>
      <div>
      </div>
    </footer>

  </body>
</html>
`

var tmplEntry = `
<!doctype html>
<html>
<meta charset="UTF-8">
  <head>
    <title>{{.Title}}</title>
    <link rel="stylesheet" type="text/css" href="../../style.css">
  </head>

  <body>
    <section>
      <article>
        {{.RenderedHTML}}
      </article>
    </section>

    <footer>
      <div>
        <a href="../../index.html">{{ .Log.Title }}</a> /
        <a href="../index.html">{{ .Group }}</a> /
        {{ .Title }}
      </div>
      <div>
        tags:
        {{ range .Tags }}
          <a href="../../{{ . }}.html">{{ . }}</a>
        {{ end }}
      </div>
      <div>
        posted on {{ FormatDate .Date }}
      </div>
    </footer>

  </body>
</html>
`
