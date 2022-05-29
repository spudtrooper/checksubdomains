package words

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	goutilio "github.com/spudtrooper/goutil/io"
)

const wordsFile = "/usr/share/dict/words"

type Checker struct {
	host    string
	timeout time.Duration
	threads int
	start   string
}

//go:generate genopts --function=New "timeout:time.Duration" "threads:int" "start:string"
func New(host string, optss ...NewOption) *Checker {
	opts := MakeNewOptions(optss...)
	return &Checker{
		host:    host,
		timeout: opts.Timeout(),
		threads: opts.Threads(),
		start:   opts.Start(),
	}
}

func (c *Checker) Check() ([]string, error) {
	lines, err := goutilio.ReadLines(wordsFile)
	if err != nil {
		return nil, err
	}
	words := make(chan string)
	go func() {
		for _, h := range lines {
			words <- h
		}
		close(words)
	}()

	type result struct {
		uri string
		ok  bool
		err error
	}

	results := make(chan result)
	go func() {
		var wg sync.WaitGroup
		start := strings.ToLower(c.start)
		for i := 0; i < c.threads; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for w := range words {
					if w == "" {
						continue
					}
					w := strings.ToLower(w)
					if start != "" && w < start {
						continue
					}
					uri := fmt.Sprintf("https://%s/%s", c.host, w)
					ok, err := c.checkURI(uri)
					if err != nil {
						log.Printf("%s %s: %v", color.HiRedString("ERROR"), uri, err)
					} else if ok {
						log.Printf("%s %s", color.GreenString("OK"), uri)
					} else {
						log.Printf("%s %s", color.RedString("NO"), uri)
					}
					res := result{
						uri: uri,
						ok:  ok,
						err: err,
					}
					results <- res
				}
			}()
		}
		wg.Wait()
		close(results)
	}()

	var uris []string
	for res := range results {
		if res.ok {
			uris = append(uris, res.uri)
		}
	}

	return uris, nil
}

func (c *Checker) checkURI(uri string) (bool, error) {
	client := &http.Client{Timeout: c.timeout}
	req, err := http.NewRequest("Get", uri, nil)
	if err != nil {
		return false, nil
	}
	resp, err := client.Do(req)
	if err != nil {
		return false, nil
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	s := string(b)
	if strings.Contains(s, "400 Bad Request") {
		return false, nil
	}

	return true, nil
}
