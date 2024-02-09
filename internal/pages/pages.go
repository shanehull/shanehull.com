package pages

import (
	"fmt"

	"context"
	"encoding/gob"
	"os"

	"github.com/gohugoio/hugo/config/allconfig"
	"github.com/gohugoio/hugo/deps"
	"github.com/gohugoio/hugo/hugofs"
	"github.com/gohugoio/hugo/hugolib"
)

type Page struct {
	Title       string
	Description string
	Content     string
	Keywords    []string
	Slug        string
}

// function that gets the pages and saves it to a gob file
func PagesToGob(contentDir string, gobDest string) error {
	// get the site
	site, err := getSite()
	if err != nil {
		return err
	}

	// get the pages
	pages, err := pagesByDir(contentDir, site)
	if err != nil {
		return err
	}

	// create a file
	dataFile, err := os.Create(gobDest)
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

func PagesFromGob(path string) ([]Page, error) {
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
