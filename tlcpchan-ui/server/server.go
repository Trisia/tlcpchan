package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Trisia/tlcpchan-ui/proxy"
)

type Server struct {
	staticDir  string
	apiProxy   *proxy.Proxy
	fileServer http.Handler
}

func New(staticDir, apiAddr string) *Server {
	absPath, err := filepath.Abs(staticDir)
	if err != nil {
		absPath = staticDir
	}

	return &Server{
		staticDir:  absPath,
		apiProxy:   proxy.New(apiAddr),
		fileServer: http.FileServer(http.Dir(absPath)),
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w, r)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/api/") {
		s.handleAPI(w, r)
		return
	}

	s.handleStatic(w, r)
}

func (s *Server) handleAPI(w http.ResponseWriter, r *http.Request) {
	s.apiProxy.ServeHTTP(w, r)
}

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(s.staticDir, r.URL.Path)

	fi, err := os.Stat(path)
	if err == nil && !fi.IsDir() {
		s.fileServer.ServeHTTP(w, r)
		return
	}

	indexPath := filepath.Join(s.staticDir, "index.html")
	if _, err := os.Stat(indexPath); err == nil {
		http.ServeFile(w, r, indexPath)
		return
	}

	http.NotFound(w, r)
}

func (s *Server) setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}

	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Max-Age", "86400")
}
