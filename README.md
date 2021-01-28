This is standard lib net/http golang realip with suport for Cloudflare etc.

# RealIP

[![GoDoc](https://godoc.org/github.com/tomasen/realip?status.svg)](http://godoc.org/github.com/tomasen/realip)

Go package that can be used to get client's real public IP, which usually useful for logging HTTP server.

### Feature

* Follows the rule of X-Real-IP
* Follows the rule of X-Forwarded-For
* Exclude local or private address


### How It Works.

It looks for specific headers in the request and falls back to some defaults if they do not exist.

The user ip is determined by the following order:

X-Client-IP
X-Original-Forwarded-For
X-Forwarded-For (Header may return multiple IP addresses in the format: "client IP, proxy 1 IP, proxy 2 IP", so we take the the first one.)
CF-Connecting-IP (Cloudflare)
Fastly-Client-Ip (Fastly CDN and Firebase hosting header when forwared to a cloud function)
True-Client-Ip (Akamai and Cloudflare)
X-Real-IP (Nginx proxy/FastCGI)
X-Forwarded, Forwarded-For and Forwarded (Variations of #2)
ctx.RemoteAddr().String()

## Example

```go
package main

import "github.com/tomasen/realip"

func (h *Handler) ServeIndexPage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	clientIP := realip.FromRequest(r)
	log.Println("GET / from", clientIP)
}
```

## Developing

Commited code must pass:

* [golint](https://github.com/golang/lint)
* [go vet](https://godoc.org/golang.org/x/tools/cmd/vet)
* [gofmt](https://golang.org/cmd/gofmt)
* [go test](https://golang.org/cmd/go/#hdr-Test_packages):
