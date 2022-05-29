package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/spudtrooper/checksubdomains/checker"
	"github.com/spudtrooper/checksubdomains/words"
	"github.com/spudtrooper/goutil/check"
)

var (
	host      = flag.String("host", "", "the host to check")
	sublist3r = flag.String("sublist3r", "", "full path to sublist3r.py")
	threads   = flag.Int("threads", 20, "number of threads for checking subdomains")
	timeout   = flag.Duration("timeout", 3*time.Second, "timeout for contacting hosts")
	outHTML   = flag.String("out_html", "", "output HTML file")
	html      = flag.Bool("html", false, "output to <html>.html; if both this and --out_html are set, --out_html wins")
	fromFile  = flag.String("from_file", "", "file containing subdomains to show")
	wordsFlag = flag.Bool("words", false, "Brute force check /usr/share/dict/words file")
	start     = flag.String("start", "", "start for the --words flag")
)

func checkWords() {
	check.Check(*host != "", check.CheckMessage("--host required"))
	c := words.New(
		*host,
		words.NewThreads(*threads),
		words.NewTimeout(*timeout),
		words.NewStart(*start),
	)
	uris, err := c.Check()
	check.Err(err)
	for u := range uris {
		fmt.Println(u)
	}
}

func checkHost() {
	check.Check(*host != "", check.CheckMessage("--host required"))
	var htmlOutputFile string
	if *outHTML != "" {
		htmlOutputFile = *outHTML
	} else if *html {
		htmlOutputFile = *host + ".html"
	}
	c := checker.New(
		*host,
		checker.NewHtmlOutputFile(htmlOutputFile),
		checker.NewSublist3r(*sublist3r),
		checker.NewThreads(*threads),
		checker.NewTimeout(*timeout),
		checker.NewSubdomainsFile(*fromFile),
	)
	subs, err := c.Check()
	check.Err(err)

	if htmlOutputFile == "" {
		for i, h := range subs {
			fmt.Printf("%4d) http://%s\n", i, h)
		}
	}
}

func main() {
	flag.Parse()
	if *wordsFlag {
		checkWords()
		return
	}
	checkHost()
}
