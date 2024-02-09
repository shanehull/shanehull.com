package main

import (
	"log"
	"os/exec"

	"github.com/shanehull/shanehull.com/internal/pages"
	"github.com/shanehull/shanehull.com/internal/util"
)

func main() {
	_, err := exec.LookPath("hugo")
	if err != nil {
		log.Fatal(
			"can't find 'hugo' in your $PATH. you probably need to install hugo: `go install -tags extended github.com/gohugoio/hugo@latest`",
		)
	}

	// build hugo site static content
	cmd := exec.Command("hugo", "--cleanDestinationDir")
	if err := util.RunCommand(cmd); err != nil {
		log.Printf("unable to build hugo site: %v", err)
	}

	// build the templ tempaltes
	cmd = exec.Command("templ", "generate")
	if err := util.RunCommand(cmd); err != nil {
		log.Printf("unable to build templ templates: %v", err)
	}

	// build the server executable
	cmd = exec.Command("go", "build", "-o", "build/server", "server.go")
	if err := util.RunCommand(cmd); err != nil {
		log.Printf("unable to build API server executable: %v", err)
	}

	// fetch pages by path and save to gob file
	if err := pages.PagesToGob("blog/", "build/pages.gob"); err != nil {
		log.Fatalf("ERROR: unable to retrieve pages from gob: %q", err)
	}

	log.Println("build complete")
}
