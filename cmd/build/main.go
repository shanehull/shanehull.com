package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/gohugoio/hugo/config/allconfig"
	"github.com/gohugoio/hugo/deps"
	"github.com/gohugoio/hugo/hugofs"
	"github.com/gohugoio/hugo/hugolib"
	"github.com/shanehull/shanehull.com/internal/helpers"
)

type Page struct {
	Title       string
	Description string
	Content     string
	Keywords    []string
	Slug        string
}

func main() {
	_, err := exec.LookPath("hugo")
	if err != nil {
		log.Fatal(
			"can't find 'hugo' in your $PATH. you probably need to install hugo: `go install -tags extended github.com/gohugoio/hugo@latest`",
		)
	}

	// build hugo site static content
	cmd := exec.Command("hugo", "--cleanDestinationDir")
	if err := helpers.RunCommand(cmd); err != nil {
		log.Printf("unable to build hugo site: %v", err)
	}

	// build the templ tempaltes
	cmd = exec.Command("templ", "generate")
	if err := helpers.RunCommand(cmd); err != nil {
		log.Printf("unable to build templ templates: %v", err)
	}

	// build the server executable
	cmd = exec.Command("go", "build", "-o", "build/server", "server.go")
	if err := helpers.RunCommand(cmd); err != nil {
		log.Printf("unable to build API server executable: %v", err)
	}

	// fetch pages by path and save to gob file
	if err := pagesToGob("pages/"); err != nil {
		log.Fatalf("ERROR: unable to retrieve pages from gob: %q", err)
	}

	log.Println("build complete")
}

// function that gets the pages and saves it to a gob file
func pagesToGob(dir string) error {
	// get the site
	site, err := getSite()
	if err != nil {
		return err
	}

	// get the pages
	pages, err := pagesByDir(dir, site)
	if err != nil {
		return err
	}

	// create a file
	dataFile, err := os.Create("build/pages.gob")
	if err != nil {
		return err
	}

	// serialize the data
	dataEncoder := gob.NewEncoder(dataFile)
	if err := dataEncoder.Encode(pages); err != nil {
		return err
	}

	dataFile.Close()

	return nil
}

func pagesByDir(dir string, site *hugolib.HugoSites) ([]Page, error) {
	pages := make([]Page, 0)

	if site != nil && (site.Pages() != nil) {
		for _, p := range site.Pages() {
			// make sure it's a page
			if p.IsPage() {
				// get the dir, it should be the same as the path
				// e.g. "blog/" for "content/blog"
				thisDir := p.File().Dir()
				if thisDir == dir {
					pageData := Page{
						Title:       p.Title(),
						Description: p.Description(),
						Content:     p.Plain(context.TODO()),
						Keywords:    p.Keywords(),
						Slug:        p.Slug(),
					}

					pages = append(pages, pageData)
				}

			}
		}
	}

	return pages, nil
}

func getSite() (*hugolib.HugoSites, error) {
	var site *hugolib.HugoSites
	osFs := hugofs.Os
	hugoPath := "./"

	config, err := allconfig.LoadConfig(allconfig.ConfigSourceDescriptor{
		Fs:          osFs,
		Filename:    "config.yaml",
		ConfigDir:   hugoPath,
		Environment: "development",
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	base := config.GetFirstLanguageConfig().BaseConfig()
	base.WorkingDir = hugoPath

	fs := hugofs.NewFrom(osFs, base)

	site, err = hugolib.NewHugoSites(deps.DepsCfg{
		Fs:      fs,
		Configs: config,
	})
	if err != nil {
		return site, err
	}

	// build (but don't render/save) so we have access to its pages, etc.
	if err := site.Build(hugolib.BuildCfg{SkipRender: true}); err != nil {
		return nil, err
	}

	return site, nil
}
