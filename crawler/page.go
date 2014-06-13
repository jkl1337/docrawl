package crawler

import (
	"net/url"
)

// Page is a single node in a site map (graph).
// XXX: Fetchers should have a separate interface or struct for their half
type Page interface {
	// URL is the URL that was used to fetch the page.
	// Any server-side redirect information will not be preserved.
	URL() *url.URL

	// Error is any error that occurred while fetching the page data.
	Error() error

	// Assets returns the collection of assets associated with the page.
	Assets() []Asset

	// Links returns all the resolved linked pages
	Links() []Page
	// GenerateLinks provides a generator for linked pages.
	GenerateLinks() <-chan Page

	// fetcher functions
	SetAssets(assets []Asset)
	SetError(err error)
}

// page is a basic non-lazy (eager) loaded page in the graph
type page struct {
	url    *url.URL
	err    error
	linked []Page
	assets []Asset
}

// newEagerPage creates a new page with empty links and assets.
func newEagerPage(u *url.URL) *page {
	// strip fragment
	if u.Fragment != "" {
		uu := *u
		uu.Fragment = ""
		u = &uu
	}
	return &page{
		url: u,
	}
}

func (p *page) URL() *url.URL {
	return p.url
}

func (p *page) Error() error {
	return p.err
}

func (p *page) SetError(err error) {
	p.err = err
}

func (p *page) Links() []Page {
	return ([]Page)(p.linked)
}

func (p *page) GenerateLinks() <-chan Page {
	pages := make(chan Page, 10)
	go func() {
		for _, v := range p.linked {
			pages <- v
		}
		close(pages)
	}()
	return pages
}

func (p *page) Assets() []Asset {
	return p.assets
}

func (p *page) SetAssets(assets []Asset) {
	p.assets = assets
}
