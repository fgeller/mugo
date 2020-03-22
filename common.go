package main

import (
	"fmt"
	"path/filepath"
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

func htmlPath(md string) string {
	bs := filepath.Base(md)
	fn := fmt.Sprintf("%s.html", bs[:len(bs)-len(".md")])
	html := filepath.Join(filepath.Dir(md), fn)
	return html
}

func verbose(fs string, args ...interface{}) {
	if true {
		fmt.Println(fmt.Sprintf(fs, args...))
	}
}

func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}
