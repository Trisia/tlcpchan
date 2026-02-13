package controller

import (
	"context"
	"net/http"
	"strings"
)

type Route struct {
	Method  string
	Pattern string
	Handler http.HandlerFunc
}

type Router struct {
	routes      []Route
	middlewares []func(http.Handler) http.Handler
}

func NewRouter() *Router {
	return &Router{
		routes:      make([]Route, 0),
		middlewares: make([]func(http.Handler) http.Handler, 0),
	}
}

func (r *Router) Use(middleware func(http.Handler) http.Handler) {
	r.middlewares = append(r.middlewares, middleware)
}

func (r *Router) Handle(method, pattern string, handler http.HandlerFunc) {
	r.routes = append(r.routes, Route{
		Method:  method,
		Pattern: pattern,
		Handler: handler,
	})
}

func (r *Router) GET(pattern string, handler http.HandlerFunc) {
	r.Handle(http.MethodGet, pattern, handler)
}

func (r *Router) POST(pattern string, handler http.HandlerFunc) {
	r.Handle(http.MethodPost, pattern, handler)
}

func (r *Router) PUT(pattern string, handler http.HandlerFunc) {
	r.Handle(http.MethodPut, pattern, handler)
}

func (r *Router) DELETE(pattern string, handler http.HandlerFunc) {
	r.Handle(http.MethodDelete, pattern, handler)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	for _, route := range r.routes {
		if route.Method != req.Method {
			continue
		}

		params, ok := matchPattern(route.Pattern, path)
		if ok {
			if len(params) > 0 {
				ctx := context.WithValue(req.Context(), pathParamsKey{}, params)
				req = req.WithContext(ctx)
			}

			handler := http.Handler(route.Handler)
			for i := len(r.middlewares) - 1; i >= 0; i-- {
				handler = r.middlewares[i](handler)
			}
			handler.ServeHTTP(w, req)
			return
		}
	}

	NotFound(w, "路由不存在")
}

type pathParamsKey struct{}

func PathParam(req *http.Request, key string) string {
	params, ok := req.Context().Value(pathParamsKey{}).(map[string]string)
	if !ok {
		return ""
	}
	return params[key]
}

func matchPattern(pattern, path string) (map[string]string, bool) {
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	if len(patternParts) != len(pathParts) {
		return nil, false
	}

	params := make(map[string]string)
	for i, pp := range patternParts {
		if strings.HasPrefix(pp, ":") {
			params[strings.TrimPrefix(pp, ":")] = pathParts[i]
		} else if pp != pathParts[i] {
			return nil, false
		}
	}

	return params, true
}
