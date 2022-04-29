package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/fatih/color"
	"github.com/spudtrooper/goutil/check"
	"github.com/spudtrooper/goutil/must"
	"github.com/spudtrooper/goutil/slice"

	_ "embed"
)

var (
	host      = flag.String("host", "", "the host to check")
	sublist3r = flag.String("sublist3r", "", "full path to sublist3r.py")
	threads   = flag.Int("threads", 20, "number of threads for checking subdomains")
	timeout   = flag.Duration("timeout", 3*time.Second, "timeout for contacting hosts")
	outHTML   = flag.String("out_html", "", "output HTML file")
	test      = flag.Bool("test", false, "use test subdomains")
	fromFile  = flag.String("from_file", "", "file containing subdomains to show")

	colorRE = regexp.MustCompile(`\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]`)
	hostRE  = regexp.MustCompile(`^[a-zA-Z0-9\-_]+(?:\.[a-zA-Z0-9\-_]+)+$`)

	//go:embed files/index.html
	indexTemplate string
	//go:embed files/jquery.js
	jqueryJS string
	//go:embed files/index.js
	indexJS string
	//go:embed files/index.css
	indexCSS string
)

func outputHTML(subdomains []string) {
	allJS := jqueryJS + "\n" + indexJS
	allCSS := indexCSS
	var data = struct {
		AllJS      string
		AllCSS     string
		Host       string
		Subdomains []string
	}{
		AllJS:      allJS,
		AllCSS:     allCSS,
		Host:       *host,
		Subdomains: subdomains,
	}
	b, err := renderTemplate(indexTemplate, "index", data)
	check.Err(err)
	must.WriteFile(*outHTML, b)
	log.Printf("wrote to %s", *outHTML)
}

func renderTemplate(t string, name string, data interface{}) ([]byte, error) {
	tmpl, err := template.New(name).Parse(strings.TrimSpace(t))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func findSubdomains() (chan string, int) {
	log.Printf("finding subdomains")

	var subdomains []string
	if *test {
		subdomains = []string{
			"foo.com",
			"bar.com",
		}
	} else {
		cmd := exec.Command("python", *sublist3r, "-d", *host)
		var stdout bytes.Buffer
		cmd.Stdout = &stdout
		check.Err(cmd.Run())
		for _, h := range slice.Strings(stdout.String(), "\n") {
			if h := strings.TrimSpace(colorRE.ReplaceAllString(h, "")); hostRE.MatchString(h) {
				subdomains = append(subdomains, h)
			}
		}
	}

	res := make(chan string)
	go func() {
		for _, h := range subdomains {
			res <- h
		}
		close(res)
	}()
	return res, len(subdomains)
}

func checkSubdomain(host string) bool {
	client := &http.Client{Timeout: *timeout}
	uri := fmt.Sprintf("http://%s", host)
	req, err := http.NewRequest("Get", uri, nil)
	if err != nil {
		return false
	}
	if _, err := client.Do(req); err != nil {
		return false
	}
	return true
}

func lookupOkSubdomains() []string {
	check.Check(*host != "", check.CheckMessage("--host required"))
	check.Check(*sublist3r != "", check.CheckMessage("--sublist3r required"))

	subDomains, numSubDomains := findSubdomains()
	log.Printf("found %d sub domains", numSubDomains)

	type result struct {
		host string
		ok   bool
	}

	results := make(chan result, numSubDomains)
	go func() {
		var wg sync.WaitGroup
		for i := 0; i < *threads; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for h := range subDomains {
					ok := checkSubdomain(h)
					results <- result{
						host: h,
						ok:   ok,
					}
					if ok {
						log.Printf("%s %s", color.GreenString("OK"), h)
					} else {
						log.Printf("%s %s", color.RedString("NO"), h)
					}
				}
			}()
		}
		wg.Wait()
		close(results)
	}()

	var subs []string
	for res := range results {
		if res.ok {
			subs = append(subs, res.host)
		}
	}

	return subs
}

func findOkSubdomains() []string {
	if *fromFile == "" {
		return lookupOkSubdomains()
	}
	return slice.NonEmptyStrings(must.ReadLines(*fromFile))
}

func checkHost() {
	subs := findOkSubdomains()
	sort.Strings(subs)

	for i, h := range subs {
		fmt.Printf("%4d) http://%s\n", i, h)
	}

	if *outHTML != "" {
		outputHTML(subs)
	}
}

func main() {
	flag.Parse()
	checkHost()
}
