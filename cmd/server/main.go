package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	root "github.com/shanehull/shanehull.com"
	"github.com/shanehull/shanehull.com/internal/buildinfo"
	"github.com/shanehull/shanehull.com/internal/handlers"
	"github.com/shanehull/shanehull.com/internal/middleware"
	"github.com/shanehull/shanehull.com/internal/pages"
)

var (
	allowedOrigin = "*"
	serverHost    = "127.0.0.1"
	serverPort    = "1314"
	logger        *slog.Logger
)

var content = &root.Public

type Page struct {
	Title       string
	Description string
	Content     string
	Keywords    []string
	Slug        string
}

func init() {
	logger = slog.New(
		slog.NewJSONHandler(
			os.Stdout, nil,
		),
	).With(slog.String("version", buildinfo.GitTag))
	slog.SetDefault(logger)
}

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(), os.Interrupt, syscall.SIGTERM)

	// Check if the SERVER_HOST env var is set and override
	envHost, ok := os.LookupEnv("SERVER_HOST")
	if ok {
		serverHost = envHost
	}

	// Check if ALLOWED_ORIGIN env var is set and override
	envOrigin, ok := os.LookupEnv("ALLOWED_ORIGIN")
	if ok {
		allowedOrigin = envOrigin
	}

	logger.Info(fmt.Sprintf("API routes allowed origin: %s", allowedOrigin))

	mux := http.NewServeMux()

	// Custom file server handler
	serverRoot, _ := fs.Sub(content, "public")
	fileServer := http.FileServer(http.FS(serverRoot))

	// Serve all hugo content (the 'public' directory) at the root url
	mux.Handle("/", fileServerWith404(fileServer, serverRoot))

	mux.HandleFunc(
		"/quote",
		middleware.CORS(
			http.HandlerFunc(
				handlers.QuoteHandler,
			),
			allowedOrigin,
		),
	)

	// TODO: endpoint for portfolio and blog search here

	// TODO: endpoint for tools here

	// Get the pages from the gob file that was generated at build time.
	// We'll use it in our search endpoint for the blog (later).
	_, err := pages.PagesFromGob("bin/pages.gob")
	if err != nil {
		log.Fatal(err)
	}

	mux.HandleFunc("/healthz", http.HandlerFunc(handlers.HealthzHandler))

	// Run the server at
	serveAt := fmt.Sprintf("%s:%s", serverHost, serverPort)
	go func() {
		if err := http.ListenAndServe(serveAt, mux); err != nil {
			log.Fatal(err)
		}
	}()

	logger.Info("API Server available", "addr", serveAt)

	// Wait for interrupt signal.
	<-ctx.Done()

	// Sleep to ensure graceful shutdown
	logger.Info("Server shutting down...")
	time.Sleep(5 * time.Second)

	// Return to default context.
	cancel()

	logger.Info("Server stopped")
}

func fileServerWith404(handler http.Handler, fs fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := fs.Open(strings.TrimPrefix(r.URL.Path, "/"))
		if os.IsNotExist(err) {
			// Rewrite request to use 404.html
			r.URL.Path = "/404.html"
		}
		handler.ServeHTTP(w, r)
	}
}
