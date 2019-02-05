package main

import (
	"flag"
	"github.com/anonymous5l/console"
	"github.com/anonymous5l/goflow/app"
	"os"
	"os/exec"
	"strconv"
)

// for main flow
func main() {
	config := flag.String("c", "flow.toml", "config path")
	pid := flag.Int("p", -1, "parent process id")

	flag.Parse()

	if *pid == -1 {
		for {
			// child agent
			cmd := exec.Command(os.Args[0], "-c", *config, "-p", strconv.Itoa(os.Getpid()))
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()

			if err != nil {
				console.Err("goflow: child process exit with error %s! ", err)
				break
			}

			console.Ok("goflow: reload child process!")
		}

		return
	}

	a, err := app.NewWebApplication(*config)

	if err != nil {
		console.Err("goflow: new application failed! %s", err)
		os.Exit(1)
		return
	}

	go func() {
		err = a.Loop()

		if err != nil {
			console.Err("goflow: %s", err)
			os.Exit(1)
			return
		}
	}()

	a.Watcher()
}
