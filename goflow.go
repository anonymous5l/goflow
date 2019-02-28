package main

import (
	"flag"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/anonymous5l/console"
	"github.com/anonymous5l/goflow/app"
)

var cmd *exec.Cmd
var mainapp *app.WebApplication

// for main flow
func main() {
	config := flag.String("c", "flow.toml", "config path")
	pid := flag.Int("p", -1, "parent process id")

	flag.Parse()

	c := make(chan os.Signal)

	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		for _ = range c {
			if *pid == -1 {
				// wait cmd processer
				// notify child processer
				if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
					console.Err("goflow: notify to child process failed! %s", err)
				}
			} else {
				mainapp.Shutdown()
			}
		}
	}()

	if *pid == -1 {
		for {
			// child agent
			cmd = exec.Command(os.Args[0], "-c", *config, "-p", strconv.Itoa(os.Getpid()))

			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Start(); err != nil {
				console.Err("goflow: failed to start command: %s", err)
			}

			if err := cmd.Wait(); err != nil {
				if exiterr, ok := err.(*exec.ExitError); ok {
					if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
						if status.ExitStatus() == -999 {
							console.Log("goflow: reload child process")
							continue
						}
					}
				}
				console.Err("goflow: failed to wait command: %s", err)
			}

			console.Ok("goflow: child done process!")
			break
		}

		return
	}

	var err error

	mainapp, err = app.NewWebApplication(*config)

	if err != nil {
		console.Err("goflow: new application failed! %s", err)
		os.Exit(1)
		return
	}

	go func() {
		err = mainapp.Loop()

		if err != nil {
			console.Err("goflow: %s", err)
			os.Exit(1)
			return
		}
	}()

	mainapp.Watcher()

	os.Exit(-999)
}
