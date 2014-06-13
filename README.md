docrawl
=======

A simple website crawler and map generator written in Golang.
This crawler demos the use of golang http client, channels, stretchr/testify, goquery,
and gographviz.

The crawler can perform concurrent http requests, but is limited to crawling a single host
name at a time. This is not the same as *same origin* as different SSL and port number
are still considered the same host.

## Usage

```shell
$ go get github.com/jkl1337/docrawl/docrawl

$ docrawl  # This will show usage
Usage: docrawl [OPTIONS] ROOT-URL
  -f="json": Output format: json: JSON, dot: Graphviz DOT, off: none
  -maxreq=2: Maximum number of simultaneous http requests
  -o="": Output filename, defaults to crawled hostname
  -pretty=false: Pretty print JSON output
  -v=false: Produce some log messages about activity

$ docrawl -v http://www.xkcd.com
2014/06/12 20:20:58 Fetching: http://www.xkcd.com/1366/
   ...
   ...
```
After the command runs successfully you should get a file www.xkcd.com.json.

To run tests make sure to do `go get -t` since stretchr/testify is a test only dependency.

## Limitations

- Conflates links and actual resolved response so redirects are not canonicalized, and
information about fragments in URLs is lost. Additionally this limitation means that
elegantly representing "external" links is not possible.
- Does not respect robots.txt. (This is not for professional web crawling use).
- The crawler API could be made lazy, but the implementation is not lazy. So early termination
of a crawl is not possible.
- Within the library not everything is fully documented.
