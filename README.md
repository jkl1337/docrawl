docrawl
=======

A tremendously simple website crawler and map generator written in Golang.

Limitations
- Conflates links and actual resolved response so redirects are not canonicalized.
- Does not respect robots.txt. (This is not for professional web crawling use).
- The crawler API could be made lazy, but the implementation is not lazy. So early termination
of a crawl is not possible.
