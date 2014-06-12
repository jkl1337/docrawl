package crawler

import (
	"testing"

	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/stretchr/testify/assert"
)

type FetchTest struct {
	responseCode int
	html         string
	assets       []string
	links        []string
	error        string
}

var fetchTests = []FetchTest{
	{
		200,
		"<html><body><a href=\"page2.html\"><img src=\"tower.jpg\"/></a></body></html>",
		[]string{"/tower.jpg"},
		[]string{"page2.html"},
		"",
	},
	{
		404,
		"<html><body><a href=\"p1\"></a><script src=\"tower.jpg\"/></script><a href=\"p2\"></a><a href=\"p3\"></a><a href=\"p4\"></a></body></html>",
		nil,
		nil,
		"non 200 status code received: 404",
	},
	{
		200,
		"<html><body><a href=\"p1\"></a><script src=\"tower.jpg\"/></script><a href=\"p2\"></a><a href=\"p3\"></a><a href=\"p4\"></a></body></html>",
		[]string{"tower.jpg"},
		[]string{"p1", "p2", "p3", "p4"},
		"",
	},
}

func TestFetchPageHTTP(t *testing.T) {

	testOne := func(tt FetchTest) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tt.responseCode)
			w.Write([]byte(tt.html))
		}))
		defer ts.Close()

		baseURL, _ := url.Parse(ts.URL)
		p := NewPage(baseURL)
		links := FetchPageHTTP(p)

		if tt.error != "" {
			assert.Error(t, p.Error(), "there should be a page fetch error")
		} else {
			assert.NoError(t, p.Error(), "there should be no page error")

			assert.Condition(t, func() bool {
				if len(tt.links) != len(links) {
					return false
				}
				for i, l := range links {
					uu, _ := baseURL.Parse(tt.links[i])
					if *(*url.URL)(l) != *uu {
						return false
					}
				}
				return true
			}, "all links should be resolved relative to base URL")

			assert.Condition(t, func() bool {
				if len(tt.assets) != len(p.Assets()) {
					return false
				}
				for i, a := range p.Assets() {
					uu, _ := baseURL.Parse(tt.assets[i])
					if *(*url.URL)(a) != *uu {
						return false
					}
				}
				return true
			}, "all assets should be resolved relative to the base URL")
		}
	}

	for _, tt := range fetchTests {
		testOne(tt)
	}
}
