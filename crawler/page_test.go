package crawler

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPage(t *testing.T) {
	u, _ := url.Parse("http://www.example.com/")
	p := NewPage(u)
	assert.Equal(t, "http://www.example.com/", p.URL().String(), "the URL is preserved")

	u, _ = url.Parse("http://www.example.com/#anchor")
	p = NewPage(u)
	assert.Equal(t, "http://www.example.com/", p.URL().String(), "the URL is preserved without fragment")
	assert.Equal(t, "http://www.example.com/#anchor", u.String(), "the input URL is not modified")
}
