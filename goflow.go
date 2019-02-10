package main

import (
	"flag"
	"github.com/anonymous5l/console"
	"github.com/anonymous5l/goflow/app"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
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
				cmd.Process.Signal(syscall.SIGQUIT)
			} else {
				mainapp.Shutdown()
			}
		}
	}()

	if *pid == -1 {
		// child agent
		cmd = exec.Command(os.Args[0], "-c", *config, "-p", strconv.Itoa(os.Getpid()))

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Start(); err != nil {
			console.Err("goflow: failed to start command: %s", err)
		}

		if err := cmd.Wait(); err != nil {
			console.Err("goflow: failed to wait command: %s", err)
		}

		console.Ok("goflow: child done process!")

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
}
