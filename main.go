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

	"github.com/fatih/color"
	"github.com/spudtrooper/goutil/check"
	"github.com/spudtrooper/goutil/slice"
)

var (
	host      = flag.String("host", "", "the host to check")
	sublist3r = flag.String("sublist3r", "", "full path to sublist3r.py")
	threads   = flag.Int("threads", 20, "number of threads for checking subdomains")
	colorRE   = regexp.MustCompile(`\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]`)
)

func findSubdomains() chan string {
	cmd := exec.Command("python", *sublist3r, "-d", *host)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	check.Err(cmd.Run())
	res := make(chan string)
	go func() {
		for _, h := range slice.Strings(stdout.String(), "\n") {
			h = colorRE.ReplaceAllString(h, "")
			h = strings.TrimSpace(h)
			if h != "" {
				res <- h
			}
		}
		close(res)
	}()
	return res
}

func checkSubdomain(host string) bool {
	if _, err := http.Get(fmt.Sprintf("http://%s", host)); err != nil {
		return false
	}
	return true
}

func checkHost() {
	check.Check(*host != "", check.CheckMessage("--host required"))
	check.Check(*sublist3r != "", check.CheckMessage("--sublist3r required"))

	subDomains := findSubdomains()
	var wg sync.WaitGroup
	results := map[string]bool{}
	for i := 0; i < *threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for h := range subDomains {
				if ok := checkSubdomain(h); ok {
					results[h] = true
					log.Printf("%s %s", color.GreenString("OK"), h)
				} else {
					results[h] = false
					log.Printf("%s %s", color.RedString("NO"), h)
				}
			}
		}()
	}
	wg.Wait()

	var okHosts []string
	for h, ok := range results {
		if ok {
			okHosts = append(okHosts, h)
		}
	}
	sort.Strings(okHosts)
	log.Printf("found %d hosts", len(okHosts))
	for i, h := range okHosts {
		fmt.Printf("%4d) http://%s\n", i, h)
	}
}

func main() {
	flag.Parse()
	checkHost()
}
