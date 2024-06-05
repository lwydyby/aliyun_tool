package server

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// SimpleServer contains necessary components of the server.
type SimpleServer struct {
	httpServer *http.Server
	fs         embed.FS
}

func buildSimpleServer(s *SimpleServer) *http.Server {
	port := os.Getenv("RUNTIME_PORT")
	if port == "" {
		port = "9010"
	}

	return &http.Server{
		Addr:    ":" + port,
		Handler: s.buildHTTPHandler(),
	}
}

func (s *SimpleServer) buildHTTPHandler() http.Handler {
	r := mux.NewRouter()
	contentStatic, _ := fs.Sub(s.fs, "front/dist")
	staticHandler := http.FileServer(http.FS(contentStatic))
	r.Path("/v1/ping").Methods(http.MethodGet).HandlerFunc(pingHandler)
	r.Path("/v1/token").Methods(http.MethodPost).Handler(CorsMiddleware(transHandler(SubmitTokenHandler)))
	r.Path("/v1/name").Methods(http.MethodPost).Handler(CorsMiddleware(transHandler(GetFileHandler)))
	r.Path("/v1/batch").Methods(http.MethodPost).Handler(CorsMiddleware(transHandler(BatchRenameHandler)))
	r.Path("/v1/dir").Methods(http.MethodPost).Handler(CorsMiddleware(transHandler(GetDirHandler)))
	r.PathPrefix("/").Handler(CorsMiddleware(http.StripPrefix("/", staticHandler)))
	return r
}

// Start starts the server.
func (s *SimpleServer) Start() {
	log.Printf("start simple server, port: %s", s.httpServer.Addr)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("server listen error: %v", err)
		}
	}()

	return
}

// Stop stops the server.
func (s *SimpleServer) Stop() {
	log.Printf("stop simple server...")

	timeout, _ := strconv.ParseInt(os.Getenv("_BYTEFAAS_FUNC_TIMEOUT"), 10, 64)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)

	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("failed to shutdown server, error: %v", err)
	}
}

// NewSimpleServer creates a new simple server.
func NewSimpleServer(fs embed.FS) (s *SimpleServer) {
	s = &SimpleServer{fs: fs}
	s.httpServer = buildSimpleServer(s)

	return
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")

		// 处理预检请求
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 继续处理请求
		next.ServeHTTP(w, r)
	})
}

func transHandler(f func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(f)
}
