package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/shanehull/shanehull.com/internal/helpers"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(), os.Interrupt, syscall.SIGTERM)

	var haveAir bool

	_, err := exec.LookPath("hugo")
	if err != nil {
		log.Fatal(
			"can't find 'hugo' in your $PATH. install hugo: `go install -tags extended github.com/gohugoio/hugo@latest`",
		)
	}

	_, err = exec.LookPath("air")
	if err == nil {
		haveAir = true
	}

	// start hugo server
	go func() {
		cmd := exec.Command("hugo", "server", "--cleanDestinationDir")
		if err := helpers.RunCommand(cmd); err != nil {
			log.Fatal(err)
		}
	}()

	// start the API server
	go func() {
		var cmd *exec.Cmd
		if haveAir {
			cmd = exec.Command("air")
		} else {
			cmd = exec.Command("go", "run", "server.go")
		}

		if err := helpers.RunCommand(cmd); err != nil {
			log.Fatal(err)
		}
	}()

	// wait for signals or context cancellation
	<-ctx.Done()

	// cancel any commands still running
	cancel()

	fmt.Println("Exited")
}
