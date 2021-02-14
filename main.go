package main

import (
	"fmt"
	"sync"
)

// Fetcher returns the body of URL and
// a slice of URLs found on that page.
type Fetcher interface {	
	Fetch(url string) (body string, urls []string, err error)
}

// URLRepo tells whether a url is already searched or not
type URLRepo struct {
	mu sync.Mutex
	urls []string
}

func (u *URLRepo) save(url string) {
	u.mu.Lock()
	u.urls = append(u.urls, url)
	u.mu.Unlock()
}

func (u *URLRepo) urlExists(url string) bool {
	u.mu.Lock()
	defer u.mu.Unlock()
	for _, item := range u.urls {
		if item == url {
			return true
		}
	}

	return false
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, repo *URLRepo, wg *sync.WaitGroup) {
	defer wg.Done()
	if depth <= 0 {
		return
	}

	if repo.urlExists(url) {
		return
	}
	
	repo.save(url)

	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("found: %s %q\n", url, body)

	for _, u := range urls {
		wg.Add(1)
		go Crawl(u, depth-1, fetcher, repo, wg)
	}

	return
}

type fakeResult struct {
	body string
	urls []string
}

type fakeFetcher map[string]*fakeResult

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}

	return "", nil, fmt.Errorf("not found: %s", url)
}

var fetcher = fakeFetcher{
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}

func main() {
	u := &URLRepo{}
	var wg sync.WaitGroup
	wg.Add(1)
	Crawl("https://golang.org/", 4, fetcher, u, &wg)
	wg.Wait()
}