package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type Proxy struct {
	target *url.URL
	proxy  *httputil.ReverseProxy
}

func New(apiAddr string) *Proxy {
	target, err := url.Parse(apiAddr)
	if err != nil {
		panic("无效的API地址: " + err.Error())
	}

	return &Proxy{
		target: target,
		proxy:  httputil.NewSingleHostReverseProxy(target),
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api")
	if r.URL.Path == "" {
		r.URL.Path = "/"
	}

	r.Host = p.target.Host
	r.Header.Set("X-Forwarded-Host", r.Host)
	r.Header.Set("X-Forwarded-Proto", "http")

	p.proxy.ServeHTTP(w, r)
}
