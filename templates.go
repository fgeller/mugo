package main

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

var tmplEntry = `
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
