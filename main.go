package main

// TODO: flags
// TODO: rss
func main() {
	var err error

	lg := newLog(
		"felix/log",
		"/Users/fgeller/src/github.com/fgeller/web/log/",
	)

	err = lg.regenerate()
	if err != nil {
		panic(err)
	}
}
