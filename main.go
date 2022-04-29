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
	"github.com/spudtrooper/goutil/io"
	"github.com/spudtrooper/goutil/must"
	"github.com/spudtrooper/goutil/slice"
)

var (
	host      = flag.String("host", "", "the host to check")
	sublist3r = flag.String("sublist3r", "", "full path to sublist3r.py")
	threads   = flag.Int("threads", 20, "number of threads for checking subdomains")
	cacheDir  = flag.String("cache_dir", ".cache", "directory of the cache")
	timeout   = flag.Duration("timeout", 3*time.Second, "timeout for contacting hosts")
	outHTML   = flag.String("out_html", "", "output HTML file")
	test      = flag.Bool("test", false, "use test subdomains")

	colorRE = regexp.MustCompile(`\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]`)
	hostRE  = regexp.MustCompile(`^[a-zA-Z0-9\-_]+(?:\.[a-zA-Z0-9\-_]+)+$`)
)

func outputHTML(subdomains []string) {
	t := `
<html>
	<head>
		<title>Subdomains for {{.Host}}</title>
	</head>
<body>
	<table style="width: 100%; height: 100%">
		<tr>
			<td style="width: 20%">
				<div style="height:100%; overflow:auto">
					<ul style="list-style:none">
					{{range .Subdomains}}
						<li>
							<a href="#" data-href="{{.}}" onclick="document.getElementById('iframe').src='http://{{.}}';">{{.}}</a>
						</li>
					{{end}}
					</ul>
				</div>
			</td>
			<td style="width: 80%">
				<iframe id="iframe" src="" style="border: 0; width: 100%; height: 100%"></iframe>
			</td>
		</tr>
	</table>
</body>
</html>
	`
	var data = struct {
		Host       string
		Subdomains []string
	}{
		Host:       *host,
		Subdomains: subdomains,
	}
	b, err := renderTemplate(t, "index", data)
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
	res := make(chan string)
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

func checkHost() {
	check.Check(*host != "", check.CheckMessage("--host required"))
	check.Check(*sublist3r != "", check.CheckMessage("--sublist3r required"))

	_, err := io.MkdirAll(*cacheDir)
	check.Err(err)

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

	var okHosts []string
	for res := range results {
		if res.ok {
			okHosts = append(okHosts, res.host)
		}
	}
	sort.Strings(okHosts)
	for i, h := range okHosts {
		fmt.Printf("%4d) http://%s\n", i, h)
	}

	if *outHTML != "" {
		outputHTML(okHosts)
	}
}

func main() {
	flag.Parse()
	checkHost()
}
