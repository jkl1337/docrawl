package docrawl

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"

	"github.com/jkl1337/docrawl/crawler"
	"github.com/stretchr/testify/assert"
)

func setupServer(dir string) *httptest.Server {
	return httptest.NewServer(http.FileServer(http.Dir(dir + "/")))
}

func testOutputResult(t *testing.T, tsurl, dir string, r *crawler.Result) {
	fn := path.Join(dir, dir+".json")

	bs, err := ioutil.ReadFile(fn)
	if err != nil {
		t.Fatalf("Unable to read test file: %s", fn)
	}
	bs = bytes.Replace(bs, []byte("http://127.0.0.1:8000"), []byte(tsurl), -1)

	var expected, actual map[string]interface{}
	if err := json.Unmarshal(bs, &expected); err != nil {
		t.Fatalf("Unable to decode test file: %s", fn)
	}

	bs, err = json.Marshal(map[string]interface{}{
		"root":  r.Root().URL().String(),
		"pages": r.LookupTable(),
	})
	assert.NoError(t, err)

	err = json.Unmarshal(bs, &actual)
	assert.NoError(t, err)

	assert.Equal(t, expected, actual)
}

func TestBroken(t *testing.T) {
	ts := setupServer("broken")
	defer ts.Close()
	c := crawler.NewCrawler(5, nil)
	r, err := c.Crawl(ts.URL)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(r.Root().Links()))

	assert.NotNil(t, r.Root().Links()[1].Error(), "expect page error for broken link")
}

func TestCircular(t *testing.T) {
	ts := setupServer("circular")
	defer ts.Close()
	c := crawler.NewCrawler(5, nil)
	r, err := c.Crawl(ts.URL)
	assert.NoError(t, err)

	testOutputResult(t, ts.URL, "circular", r)
}
