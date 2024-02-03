package main

import (
	"bytes"
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
)

//go:embed all:public
var content embed.FS

var (
	serverProto string
	serverHost  string
	serverPort  string
	serverAddr  string
)

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

	// Default server address
	serverProto, serverHost, serverPort = "http", "localhost", "1314"
	serverAddr = fmt.Sprintf("%s://%s:%s", serverProto, serverHost, serverPort)

	// Check if the SERVER_HOST env var is set
	envHost, ok := os.LookupEnv("SH_SERVER_HOST")
	if ok {
		// If it is, use it as the server host (cors)
		serverProto, serverHost, serverPort = "https", envHost, "1314"
		serverAddr = fmt.Sprintf("%s://%s", serverProto, serverHost)
	}

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

	mux.HandleFunc("/hello_world", cors(http.HandlerFunc(helloWorld)))

	// TODO: endpoint for blog search server
	// TODO: endpoint for quotes api server

	go func() {
		serveAt := fmt.Sprintf("%s:%s", serverHost, serverPort)

		if err := http.ListenAndServe(serveAt, mux); err != nil {
			log.Fatal(err)
		}
	}()

	fmt.Printf("API Server available at %s\n", serverAddr)

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
		w.Header().Set("Access-Control-Allow-Origin", serverAddr)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST")
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

// It responds with the the HTML partial `partials/helloworld.html`
func helloWorld(w http.ResponseWriter, r *http.Request) {
	buf := &bytes.Buffer{}

	w.WriteHeader(http.StatusOK)

	_, err := w.Write(buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}
}
