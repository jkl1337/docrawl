package crawler

import (
	"net/url"
)

// Page is a single node in a site map (graph).
type Page struct {
	url    *url.URL
	err    error
	linked []*Page
	assets []Asset
}

// URL is the URL that was used to fetch the page.
// Any server-side redirect information will not be preserved.
func (p *Page) URL() *url.URL {
	return p.url
}

// Error is any error that occurred while fetching the page data.
// If Error is non-nil, then the rest of the page fields are undefined.
func (p *Page) Error() error {
	return p.err
}

func (p *Page) SetError(err error) {
	p.err = err
}

func (p *Page) AllLinks() []*Page {
	return p.linked
}

// Assets returns the collection of assets associated with the page.
func (p *Page) Assets() []Asset {
	return p.assets
}

func (p *Page) SetAssets(assets []Asset) {
	p.assets = assets
}

// NewPage creates a new page with empty links and assets.
func NewPage(u *url.URL) *Page {
	// strip fragment
	if u.Fragment != "" {
		uu := *u
		uu.Fragment = ""
		u = &uu
	}
	return &Page{
		url: u,
	}
}
