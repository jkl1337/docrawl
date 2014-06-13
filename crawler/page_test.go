package crawler

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

// testPageLinks is a generic tester for any Page implementation
func testPageLinks(t *testing.T, setup func() (p Page, expected []Page)) {
	var p Page
	var expected []Page

	resolveGenerated := func(p Page) (g []Page) {
		g = make([]Page, 0)
		for p := range p.GenerateLinks() {
			g = append(g, p)
		}
		return g
	}

	p, expected = setup()
	assert.Equal(t, expected, resolveGenerated(p), "GenerateLinks() produces all links")
	assert.Equal(t, expected, resolveGenerated(p), "GenerateLinks() is idempotent")

	p, expected = setup()
	assert.Equal(t, expected, p.Links(), "Links() returns all links")
	assert.Equal(t, expected, resolveGenerated(p), "GenerateLinks() is still idempotent")
}

func TestNewPage(t *testing.T) {
	u, _ := url.Parse("http://www.example.com/")
	p := newEagerPage(u)
	assert.Equal(t, "http://www.example.com/", p.URL().String(), "the URL is preserved")

	u, _ = url.Parse("http://www.example.com/#anchor")
	p = newEagerPage(u)
	assert.Equal(t, "http://www.example.com/", p.URL().String(), "the URL is preserved without fragment")
	assert.Equal(t, "http://www.example.com/#anchor", u.String(), "the input URL is not modified")
}

func TestSimplePageLinks(t *testing.T) {
	testPageLinks(t, func() (Page, []Page) {
		u, _ := url.Parse("http://www.example.com/")
		p := newEagerPage(u)
		p.linked = make([]Page, 100)
		for i := 0; i < len(p.linked); i++ {
			p.linked[i] = &page{}
		}
		return p, p.linked
	})
}
