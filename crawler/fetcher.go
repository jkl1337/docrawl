package crawler

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

// FetchPageHTTP is a simple http only crawler fetcher. It populates the Page Assets by
// scraping the page with the standard golang HTML parser.
func FetchPageHTTP(p *Page) []*url.URL {
	res, err := http.Get(p.URL().String())
	if err != nil {
		p.SetError(err)
		return nil
	}
	if res.StatusCode != 200 {
		p.SetError(fmt.Errorf("non 200 status code received: %v", res.StatusCode))
		return nil
	}

	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		p.SetError(err)
		return nil
	}
	p.SetError(nil)

	links := make([]*url.URL, 0, 8)
	doc.Find("a[href]").Each(func(n int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if len(href) == 0 {
			return
		}
		u, err := url.Parse(href)
		if err != nil {
			return
		}
		if len(u.Scheme) > 0 && u.Scheme != "http" && u.Scheme != "https" {
			return
		}
		u = p.URL().ResolveReference(u)

		if u.Host == p.URL().Host {
			links = append(links, u)
		}
	})

	assets := make([]Asset, 0)
	assetSels := map[string]string{
		"img[src]":    "src",
		"script[src]": "src",
		"link[href]":  "href",
	}
	for selStr, attr := range assetSels {
		doc.Find(selStr).Each(func(n int, s *goquery.Selection) {
			src, _ := s.Attr(attr)
			if len(src) == 0 {
				return
			}
			assetURL, _ := p.URL().Parse(src)
			if assetURL != nil {
				assets = append(assets, assetURL)
			}
		})
	}
	p.SetAssets(assets)
	return links
}
