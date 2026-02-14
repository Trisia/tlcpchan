package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Trisia/tlcpchan-ui/proxy"
)

type Server struct {
	staticDir  string
	apiProxy   *proxy.Proxy
	fileServer http.Handler
	version    string
}

type versionResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Version   string `json:"version"`
		GoVersion string `json:"go_version"`
	} `json:"data"`
}

func New(staticDir, apiAddr, version string) *Server {
	absPath, err := filepath.Abs(staticDir)
	if err != nil {
		absPath = staticDir
	}

	return &Server{
		staticDir:  absPath,
		apiProxy:   proxy.New(apiAddr),
		fileServer: http.FileServer(http.Dir(absPath)),
		version:    version,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w, r)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.URL.Path == "/api/v1/ui/version" {
		s.versionHandler(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/api/") {
		s.apiProxy.ServeHTTP(w, r)
		return
	}

	s.handleStatic(w, r)
}

func (s *Server) versionHandler(w http.ResponseWriter, r *http.Request) {
	resp := versionResponse{
		Code:    0,
		Message: "success",
	}
	resp.Data.Version = s.version
	resp.Data.GoVersion = runtime.Version()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
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
