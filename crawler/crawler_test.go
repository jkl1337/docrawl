package crawler

import (
	"fmt"
	"net/url"
	"runtime"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mapURLs(b *url.URL, raw []string) (urls []*url.URL) {
	urls = make([]*url.URL, len(raw))
	for i, s := range raw {
		var e error
		if b != nil {
			urls[i], e = b.Parse(s)
		} else {
			urls[i], e = url.Parse(s)
		}
		if e != nil {
			panic(fmt.Errorf("parse error in test URL %v, %v", s, e))
		}
	}
	return
}

func TestnewPageMap(t *testing.T) {
	pm := newPageMap("")
	assert.NotNil(t, pm.pages)
}

func TestpageMap_getPages(t *testing.T) {
	links := mapURLs(nil, []string{
		"http://testhost.local/page1.html",
		"http://testhost.local/page2.html?query3#anchor",
		"/page3.html",
	})

	pm := newPageMap("testhost.local")
	pages, newPages := pm.getPages(links)

	// per spec, URLs must be absolute
	if assert.Equal(t, 2, len(pages), "all same host pages should be returned") {
		assert.Equal(t, *links[0], *pages[0].URL(), "URL should be preserved")
	}
	assert.Equal(t, 2, len(newPages), "all pages should be new")

	links = mapURLs(nil, []string{
		"http://testhost.local/page1.html",
	})
	pages, newPages = pm.getPages(links)

	if assert.Equal(t, 1, len(pages), "all requested pages should be returned") {
		assert.Equal(t, *links[0], *pages[0].URL(), "URL should be preserved")
	}
	assert.Equal(t, 0, len(newPages), "existing pages are not returned new")

	links = mapURLs(nil, []string{
		"http://testhost.local/page2.html?query3",
	})
	pages, newPages = pm.getPages(links)

	assert.Equal(t, 1, len(pages), "all requested pages should be returned")
	assert.Equal(t, 0, len(newPages), "pages with same HTTP request URI are not returned new")
}

var pages = map[string][]string{
	"/":            {"/page1.html", "/page2.html", "/page3.html"},
	"/page1.html":  {"/page2.html", "/page3.html"},
	"/page2.html":  {"/page2.html", "/page3.html"},
	"/page3.html":  {"/page4.html", ""},
	"/page4.html":  {"/page5.html", "/page1.html"},
	"/page5.html":  {"/page2.html", "/page6.html"},
	"/page6.html":  {"/page7.html", "/page8.html", "/page9.html", "/page10.html", "/page11.html", "/page12.html"},
	"/page7.html":  {},
	"/page8.html":  {},
	"/page9.html":  {},
	"/page10.html": {},
	"/page11.html": {},
	"/page12.html": {},
}

func TestCrawlerErrors(t *testing.T) {
	var numFetches uint32

	fetcher := func(p *Page) []*url.URL {
		atomic.AddUint32(&numFetches, 1)
		return nil
	}

	c := NewCrawler(0, fetcher)
	_, err := c.Crawl("%gh&%ij")

	assert.Error(t, err, "error is returned for malformed root URL")
	assert.Equal(t, 0, numFetches, "no fetches occur for root error")
}

func TestCrawler(t *testing.T) {
	baseURL, _ := url.Parse("http://testhost.local/")

	var numFetches uint32
	fetcher := func(p *Page) []*url.URL {
		atomic.AddUint32(&numFetches, 1)
		links := pages[p.URL().RequestURI()]
		if links == nil {
			t.Errorf("test requesting nonexistant URI: %v", p.URL().RequestURI())
		}
		return mapURLs(p.URL(), links)
	}

	// try a few different concurrency settings
	for _, nc := range []int{0, 1, 2, 50000} {
		atomic.StoreUint32(&numFetches, 0)

		c := NewCrawler(nc, fetcher)
		cr, err := c.Crawl(baseURL.String())

		assert.NoError(t, err)
		assert.Equal(t, len(pages), numFetches, "all pages were fetched")
		assert.Equal(t, len(pages), len(cr.LookupTable()), "the lookup table is complete")
		assert.Equal(t, baseURL.String(), cr.Root().URL().String(), "the root URL is corect")

		visited := map[*Page]bool{}
		var cmpTree func(p *Page)
		cmpTree = func(p *Page) {
			visited[p] = true
			expected := pages[p.URL().RequestURI()]
			assert.NotNil(t, expected)

			for i, lp := range p.AllLinks() {
				assert.Equal(t, expected[i], lp.URL().RequestURI())
				if !visited[lp] {
					cmpTree(lp)
				}
			}
		}
	}
}

func TestCrawlerConcurrency(t *testing.T) {
	var numFetches uint32
	var numConcurrent, maxConcurrent int32
	baseURL, _ := url.Parse("http://testhost.local/")

	fetcher := func(p *Page) []*url.URL {
		cur := atomic.AddInt32(&numConcurrent, 1)
		for {
			old := atomic.SwapInt32(&maxConcurrent, cur)
			if old <= cur {
				break
			}
			cur = old
		}
		defer atomic.AddInt32(&numConcurrent, -1)

		runtime.Gosched()
		atomic.AddUint32(&numFetches, 1)
		links := pages[p.URL().RequestURI()]
		if links == nil {
			t.Errorf("test requesting nonexistant URI: %v", p.URL().RequestURI())
		}
		return mapURLs(p.URL(), links)
	}

	c := NewCrawler(3, fetcher)
	c.Crawl(baseURL.String())

	assert.Equal(t, len(pages), numFetches, "all pages were fetched")
	assert.Condition(t, func() bool {
		return !(maxConcurrent > 3 || maxConcurrent < 2)
	}, "maximum concurrency outside of specification (2 <= %v <= 3)", maxConcurrent)

}
