package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	gv "code.google.com/p/gographviz"
	"github.com/jkl1337/docrawl/crawler"
)

var (
	verbose      = flag.Bool("v", false, "Produce some log messages about activity")
	maxRequests  = flag.Int("maxreq", 2, "Maximum number of simultaneous http requests")
	outputFormat = flag.String("f", "json", "Output format: json: JSON, dot: Graphviz DOT, off: none")
	pretty       = flag.Bool("pretty", false, "Pretty print JSON output")
	outputName   = flag.String("o", "", "Output filename, defaults to crawled hostname")
)

type ResultWriter interface {
	Ext() string
	Write(w io.Writer, cr *crawler.Result) error
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] ROOT-URL\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	var serializer ResultWriter
	switch *outputFormat {
	case "json":
		serializer = jsonWriter{}
	case "dot":
		serializer = dotWriter{}
	case "off":
	default:
		fmt.Fprintf(os.Stderr, "Invalid output format")
		os.Exit(2)
	}

	rooturl := flag.Arg(0)
	if rooturl == "" {
		flag.Usage()
		os.Exit(2)
	}

	fetcher := crawler.FetchPageHTTP
	if *verbose {
		fetcher = func(p *crawler.Page) []*url.URL {
			log.Println("Fetching:", p.URL().String())
			return crawler.FetchPageHTTP(p)
		}
	}

	c := crawler.NewCrawler(*maxRequests, fetcher)
	cr, err := c.Crawl(rooturl)

	if err != nil {
		log.Fatalln("Crawler failed", err)
	}

	if serializer != nil {
		name := *outputName
		if name == "" {
			name = fmt.Sprintf("%s.%s", cr.Root().URL().Host, serializer.Ext())
		}
		var f io.Writer
		if name == "-" {
			f = os.Stdout
		} else {
			osf, err := os.Create(name)
			defer osf.Close()
			f = osf
			if err != nil {
				log.Fatalf("Unable to open output file: %s, %v", name, err)
			}
		}
		if err = serializer.Write(f, cr); err != nil {
			log.Fatalf("Unable to open output file: %s, %v", name, err)
		}
	}
}

type jsonWriter struct{}

func (j jsonWriter) Ext() string {
	return "json"
}

func (j jsonWriter) Write(w io.Writer, cr *crawler.Result) error {
	var err error
	var bs []byte
	if *pretty {
		bs, err = json.MarshalIndent(cr.LookupTable(), "", "  ")
	} else {
		bs, err = json.Marshal(cr.LookupTable())
	}
	if err != nil {
		return err
	}
	_, err = w.Write(bs)
	return err
}

type dotWriter struct{}

func (j dotWriter) Ext() string {
	return "dot"
}

func nodeLabel(p *crawler.Page) string {
	abuf := make([]string, 0)
	for _, a := range p.Assets() {
		abuf = append(abuf, (*url.URL)(a).String())
	}
	abuf = append(abuf, "")
	assetsStr := strings.Join(abuf, "\\l")
	return fmt.Sprintf("{%s|%s}", p.URL().String(), assetsStr)
}

func (j dotWriter) Write(w io.Writer, cr *crawler.Result) error {
	var err error
	name := cr.Root().URL().Host

	g := gv.NewEscape()
	g.SetDir(true)
	g.SetName(name)

	labelCount := 1
	visited := map[*crawler.Page]int{}

	pageID := func(p *crawler.Page) string {
		return "P" + strconv.FormatInt(int64(visited[p]), 10)
	}

	var walkPage func(p *crawler.Page)
	walkPage = func(p *crawler.Page) {
		nodeAttrs := map[string]string{
			"shape": "record",
			"label": nodeLabel(p),
		}
		g.AddNode(name, pageID(p), nodeAttrs)

		edges := map[*crawler.Page]int{}

		for _, lp := range p.AllLinks() {
			edges[lp]++
			if visited[lp] == 0 {
				visited[lp] = labelCount
				labelCount++
				walkPage(lp)
			}
		}
		for lp, w := range edges {
			var edgeAttrs map[string]string
			if w > 1 {
				edgeAttrs = map[string]string{
					"label": strconv.FormatInt(int64(w), 10),
				}
			}
			g.AddEdge(pageID(p), pageID(lp), true, edgeAttrs)
		}
	}
	r := cr.Root()
	visited[r] = labelCount
	labelCount++
	walkPage(r)

	// TODO: graphviz panics for errors, so trap it here
	_, err = w.Write([]byte(g.String()))
	return err
}
