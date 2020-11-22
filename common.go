package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

func measure(f func() error, eh func(error), mf string, args ...interface{}) {
	start := time.Now()
	err := f()
	if err != nil {
		eh(err)
	}
	elapsed := time.Since(start)
	args = append(args, elapsed.Milliseconds())
	verbose(mf, args...)
}

func fail(err error) {
	if err != nil {
		panic(err)
	}
}

func urlJoin(args ...string) string {
	result := args[0]
	for i, u := range args {
		if i == 0 {
			continue
		}
		if !strings.HasSuffix(result, "/") {
			result += "/"
		}
		result += u
	}
	return result
}

func verbose(fs string, args ...interface{}) {
	if true {
		fmt.Println(fmt.Sprintf(fs, args...))
	}
}

func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func TimeLayout(t time.Time, l string) string {
	return t.Format(l)
}

func inferHTMLFilePath(b *blog, md string) (string, error) {
	rel, err := filepath.Rel(b.BaseDirectory, md)
	if err != nil {
		return "", err
	}

	bs := filepath.Base(rel)
	fn := fmt.Sprintf("%s.html", bs[:len(bs)-len(".md")])
	out := filepath.Join(b.OutputDirectory, filepath.Dir(rel), fn)

	return out, nil
}
