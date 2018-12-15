package main

import (
	"fmt"
	"github.com/anonymous5l/console"
	"github.com/valyala/fasthttp"
	"goflow/cfg"
	"goflow/impl"
	"net"
	"os"
)

var coreCtx *impl.ContextImpl

func getListener(t string, addr string) (net.Listener, error) {
	if t == "unix" {
		if err := os.Remove(addr); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("fasthttp: unexpected error when trying to remove unix socket file %q: %s", addr, err)
		}
	}
	ln, err := net.Listen(t, addr)
	if err != nil {
		return nil, err
	}
	if t == "unix" {
		if err = os.Chmod(addr, os.ModePerm); err != nil {
			return nil, fmt.Errorf("fasthttp: cannot chmod %#o for %q: %s", os.ModePerm, addr, err)
		}
	}
	return ln, nil
}

func requestHandle(ctx *fasthttp.RequestCtx) {
	coreCtx.Handle(ctx)
}

// for main flow
func main() {
	var err error

	args := os.Args
	argc := len(args)

	config := "flow.toml"

	if argc > 1 {
		config = args[1]
	}

	coreCtx, err = cfg.InitConfig(config)

	if err != nil {
		console.Err("goflow: load config %s", err)
		return
	}

	ln, err := getListener(cfg.Listen.Type, cfg.Listen.Addr)

	if err != nil {
		console.Err("goflow: getListener %s", err)
		return
	}

	console.Ok("goflow: running on %s %s", cfg.Listen.Type, cfg.Listen.Addr)

	fasthttp.Serve(ln, requestHandle)

	if err != nil {
		console.Err("fasthttp: %s", err)
		return
	}
}
