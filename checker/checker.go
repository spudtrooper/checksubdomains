package checker

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/spudtrooper/goutil/io"
	goutillog "github.com/spudtrooper/goutil/log"
	"github.com/spudtrooper/goutil/or"
	"github.com/spudtrooper/goutil/slice"

	_ "embed"
)

var (
	colorRE = regexp.MustCompile(`\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]`)
	hostRE  = regexp.MustCompile(`^[a-zA-Z0-9\-_]+(?:\.[a-zA-Z0-9\-_]+)+$`)

	log = goutillog.MakeLog("checksubdomains", goutillog.MakeLogColor(true))

	//go:embed files/index.html
	indexTemplate string
	//go:embed files/jquery.js
	jqueryJS string
	//go:embed files/index.js
	indexJS string
	//go:embed files/index.css
	indexCSS string
)

type Checker struct {
	host           string
	sublist3r      string
	timeout        time.Duration
	threads        int
	subdomainsFile string
	htmlOutputFile string
	verbose        bool
}

func (c *Checker) log(tmpl string, args ...interface{}) {
	if c.verbose {
		log.Printf(tmpl, args...)
	}
}

//go:generate genopts --function=New "sublist3r:string" "timeout:time.Duration" "threads:int" "subdomainsFile:string" "htmlOutputFile:string" "verbose"
func New(host string, optss ...NewOption) *Checker {
	opts := MakeNewOptions(optss...)
	return &Checker{
		host:           host,
		sublist3r:      opts.Sublist3r(),
		timeout:        opts.Timeout(),
		threads:        opts.Threads(),
		subdomainsFile: opts.SubdomainsFile(),
		htmlOutputFile: opts.HtmlOutputFile(),
		verbose:        opts.Verbose(),
	}
}

func (c *Checker) Check() ([]string, error) {
	subs, err := c.findOkSubdomains()
	if err != nil {
		return nil, err
	}
	sort.Strings(subs)

	if f := c.htmlOutputFile; f != "" {
		if err := c.outputHTML(f, subs); err != nil {
			return nil, err
		}
	}

	return subs, nil
}

func (c *Checker) outputHTML(outputFile string, subdomains []string) error {
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
		Host:       c.host,
		Subdomains: subdomains,
	}
	b, err := renderTemplate(indexTemplate, "index", data)
	if err != nil {
		return err
	}
	if err := io.WriteFile(outputFile, b); err != nil {
		return err
	}
	abs, err := filepath.Abs(outputFile)
	if err != nil {
		return err
	}
	f := fmt.Sprintf("file://%s", abs)
	log.Printf("wrote to %s", f)
	return nil
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

func (c *Checker) findSubdomains() (chan string, int, error) {
	log.Printf("finding subdomains for %s", "http://"+c.host)

	var subdomains []string

	sublist3rPY := or.String(c.sublist3r, os.Getenv("SUBLIST3R_PY"))
	if sublist3rPY == "" {
		return nil, 0, errors.Errorf("set either --sublist3r or the SUBLIST3R_PY env variable")
	}
	cmd := exec.Command("python", sublist3rPY, "-d", c.host)
	c.log("command line: %v", cmd)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, 0, err
	}
	for _, h := range slice.Strings(stdout.String(), "\n") {
		if h := strings.TrimSpace(colorRE.ReplaceAllString(h, "")); hostRE.MatchString(h) {
			subdomains = append(subdomains, h)
		}
	}

	res := make(chan string)
	go func() {
		for _, h := range subdomains {
			res <- h
		}
		close(res)
	}()
	return res, len(subdomains), nil
}

func (c *Checker) checkSubdomain(host string) bool {
	client := &http.Client{Timeout: c.timeout}
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

func (c *Checker) lookupOkSubdomains() ([]string, error) {
	if c.host == "" {
		return nil, errors.Errorf("--host required")
	}

	subDomains, numSubDomains, err := c.findSubdomains()
	if err != nil {
		return nil, err
	}
	log.Printf("found %d sub domains", numSubDomains)

	type result struct {
		host string
		ok   bool
	}

	results := make(chan result, numSubDomains)
	go func() {
		var wg sync.WaitGroup
		for i := 0; i < c.threads; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for h := range subDomains {
					ok := c.checkSubdomain(h)
					results <- result{
						host: h,
						ok:   ok,
					}
					uri := fmt.Sprintf("http://%s", h)
					if ok {
						log.Printf("%s %s", color.GreenString("OK"), uri)
					} else {
						log.Printf("%s %s", color.RedString("NO"), uri)
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

	return subs, nil
}

func (c *Checker) findOkSubdomains() ([]string, error) {
	if c.subdomainsFile == "" {
		return c.lookupOkSubdomains()
	}
	lines, err := io.ReadLines(c.subdomainsFile)
	if err != nil {
		return nil, err
	}
	res := slice.NonEmptyStrings(lines)
	return res, nil
}
