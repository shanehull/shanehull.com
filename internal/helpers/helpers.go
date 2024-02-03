package helpers

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
)

func RunCommand(cmd *exec.Cmd) error {
	outPipe, _ := cmd.StdoutPipe()
	errPipe, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		log.Fatalf("could not run %s: %v", cmd.String(), err)
	}

	go func() {
		// print the output from the command in real time
		scanner := bufio.NewScanner(outPipe)
		for scanner.Scan() {
			m := scanner.Text()
			fmt.Println(m)
		}
	}()

	go func() {
		// print err output from the command in real time
		scanner := bufio.NewScanner(errPipe)
		for scanner.Scan() {
			m := scanner.Text()
			fmt.Println(m)
		}
	}()

	if err := cmd.Wait(); err != nil {
		log.Fatalf("ERROR: cmd %s stopped: %q", cmd.String(), err)
	}

	return nil
}
