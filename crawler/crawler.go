package crawler

import (
	"net/url"
	"sync"
)

const defaultMaxRequests = 2

// Fetcher is a function type that can fetch and fill out Page structures by some means.
// The return value is slice of links. The crawler will then build the links in the page at
// some time.
type Fetcher func(p Page) []*url.URL

// Asset is a URL to a page asset (img, css, script)
type Asset *url.URL

// Crawler is a basic single-domain website crawler.
type Crawler struct {
	maxRequests int
	fetcher     Fetcher
}

// PageRecord is a marshalable record of a page with references only by string.
type PageRecord struct {
	Links  []string `json:"links,omitempty"`
	Assets []string `json:"assets,omitempty"`
	Error  string   `json:"error,omitempty"`
}

// Result provides access to the result of a crawl.
type Result struct {
	root   Page
	pages  map[string]Page
	lookup map[string]PageRecord
}

// Root returns the root page for the crawl.
func (cr *Result) Root() Page {
	return cr.root
}

// LookupTable returns a page map/table that is suitable for serialization.
func (cr *Result) LookupTable() map[string]PageRecord {
	if cr.lookup != nil {
		return cr.lookup
	}
	cr.lookup = map[string]PageRecord{}
	for _, p := range cr.pages {
		pr := PageRecord{}
		if p.Error() == nil {
			pr.Links = make([]string, len(p.Links()))
			pr.Assets = make([]string, len(p.Assets()))

			for i, a := range p.Assets() {
				pr.Assets[i] = (*url.URL)(a).String()
			}
			for i, l := range p.Links() {
				pr.Links[i] = l.URL().String()
			}
		} else {
			pr.Error = p.Error().Error()
		}
		cr.lookup[p.URL().String()] = pr
	}
	cr.pages = nil
	return cr.lookup
}

type sentinel struct{}

// newPageMap is visited page memo that uses the HTTP request URI as key
// and limits the map to only pages on a single host.
type pageMap struct {
	host  string
	lock  sync.Mutex
	pages map[string]Page
}

func newPageMap(host string) *pageMap {
	return &pageMap{
		host:  host,
		pages: make(map[string]Page),
	}
}

// getPages returns Page structs for all given links that are within the
// host associated with this pageMap. The first returned slice is the
// same length as links and contains a page instance for every link. The
// second slice returned is a subset of the elements of the first slice,
// containing all newly initialized pages.
func (pm *pageMap) getPages(links []*url.URL) ([]Page, []Page) {
	keys := make([]string, 0, len(links))
	for _, l := range links {
		if l.Host == pm.host {
			keys = append(keys, l.RequestURI())
		}
	}

	pages := make([]Page, len(keys))
	newPages := make([]Page, 0)

	pm.lock.Lock()
	defer pm.lock.Unlock()
	for i, k := range keys {
		page := pm.pages[k]
		if page == nil {
			page = newEagerPage(links[i])
			pm.pages[k] = page
			newPages = append(newPages, page)
		}
		pages[i] = page
	}
	return pages, newPages
}

func NewCrawler(maxRequests int, fetcher Fetcher) *Crawler {
	if fetcher == nil {
		fetcher = FetchPageHTTP
	}
	if maxRequests == 0 {
		maxRequests = defaultMaxRequests
	}
	return &Crawler{
		maxRequests: maxRequests,
		fetcher:     fetcher,
	}
}

type crawlerState struct {
	fetchSemaphore chan sentinel
	wg             sync.WaitGroup
	fetcher        Fetcher
	pageMap        *pageMap
}

// Crawl synchronously crawls the rootURL for links within the same host
// using the fetcher.
func (c *Crawler) Crawl(rootURL string) (*Result, error) {
	u, err := url.Parse(rootURL)
	if err != nil {
		return nil, err
	}

	cs := &crawlerState{
		fetchSemaphore: make(chan sentinel, c.maxRequests),
		pageMap:        newPageMap(u.Host),
		fetcher:        c.fetcher,
	}
	for i := 0; i < c.maxRequests; i++ {
		cs.fetchSemaphore <- sentinel{}
	}

	rootPage := newEagerPage(u)
	cs.pageMap.pages[u.RequestURI()] = rootPage
	cs.fetchPage(rootPage)

	cs.wg.Wait()
	return &Result{
		root:  rootPage,
		pages: cs.pageMap.pages,
	}, nil
}

func (cs *crawlerState) fetchPage(p Page) {
	cs.wg.Add(1)
	<-cs.fetchSemaphore
	go func() {
		links := cs.fetcher(p)

		cs.fetchSemaphore <- sentinel{}

		if len(links) >= 1 {
			linked, unfetched := cs.pageMap.getPages(links)
			p.(*page).linked = linked
			for _, np := range unfetched {
				cs.fetchPage(np)
			}
		}
		cs.wg.Done()
	}()
}
