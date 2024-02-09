package main

import (
	"context"
	"embed"
	"encoding/gob"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shanehull/shanehull.com/api"
)

//go:embed all:public
var content embed.FS

var allowedOrigin = "*"
var serverHost = "127.0.0.1"
var serverPort = "1314"

type Page struct {
	Title       string
	Description string
	Content     string
	Keywords    []string
	Slug        string
}

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(), os.Interrupt, syscall.SIGTERM)

	envHost, ok := os.LookupEnv("SH_SERVER_HOST")
	if ok {
		serverHost = envHost
	}

	// Check if the SERVER_HOST env var is set
	envOrigin, ok := os.LookupEnv("SH_ALLOW_ORIGIN")
	if ok {
		allowedOrigin = envOrigin
	}
	fmt.Printf("Allowed origin: %s\n", allowedOrigin)

	mux := http.NewServeMux()

	serverRoot, _ := fs.Sub(content, "public")

	// Serve all hugo content (the 'public' directory) at the root url
	mux.Handle("/", http.FileServer(http.FS(serverRoot)))

	// Get the pages from the gob file that was generated at build time.
	// We'll use it in our search endpoint for the blog.
	_, err := getPagesFromGob("build/pages.gob")
	if err != nil {
		log.Fatal(err)
	}

	mux.HandleFunc("/quote", cors(http.HandlerFunc(api.QuoteHandler)))
	// TODO: endpoint for portfolio and blog search

	serveAt := fmt.Sprintf("%s:%s", serverHost, serverPort)

	go func() {
		if err := http.ListenAndServe(serveAt, mux); err != nil {
			log.Fatal(err)
		}
	}()

	fmt.Printf("API Server available at %s\n", serveAt)

	// Wait for interrupt signal.
	<-ctx.Done()

	// Sleep to ensure graceful shutdown
	fmt.Println("Server shutting down...")
	time.Sleep(5 * time.Second)

	// Return to default context.
	cancel()

	fmt.Println("Server stopped")
}

func cors(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers",
			"Content-Type, hx-target, hx-current-url, hx-request")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		h.ServeHTTP(w, r)
	}
}

func getPagesFromGob(path string) ([]Page, error) {
	var pages []Page

	f, err := os.Open(path)
	if err != nil {
		return pages, err
	}

	if err := gob.NewDecoder(f).Decode(&pages); err != nil {
		return pages, err
	}

	f.Close()

	return pages, nil
}
