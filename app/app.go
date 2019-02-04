package app

import (
	"fmt"
	"github.com/anonymous5l/console"
	"github.com/anonymous5l/goflow/cfg"
	"github.com/anonymous5l/goflow/impl"
	"github.com/fsnotify/fsnotify"
	"github.com/valyala/fasthttp"
	"net"
	"os"
	"path/filepath"
	"strings"
)

type WebApplication struct {
	Path        string
	Config      *cfg.Config
	Context     *impl.ContextImpl
	Listener    net.Listener
	watcher     *fsnotify.Watcher
	fileWatcher []*cfg.HashFile
}

func (self *WebApplication) handle(ctx *fasthttp.RequestCtx) {
	self.Context.Handle(ctx)
}

func (self *WebApplication) addToWatcher(path string) error {
	if p, err := filepath.Abs(path); err != nil {
		return err
	} else if err := self.watcher.Add(filepath.Dir(p)); err != nil {
		return err
	} else if fh, err := cfg.NewHashFile(p); err != nil {
		return err
	} else {
		// console.Debug("add to watcher %s", p)
		self.fileWatcher = append(self.fileWatcher, fh)
	}

	return nil
}

func (self *WebApplication) Watcher() error {
	if self.watcher != nil {
		self.watcher.Close()
	}

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		return err
	}

	self.watcher = watcher

	err = self.addToWatcher(self.Path)

	if err != nil {
		return err
	}

	for _, p := range self.Config.Services {
		err = self.addToWatcher(p.Path)

		if err != nil {
			return err
		}
	}

	for _, p := range self.Config.Flow {
		err = self.addToWatcher(p.Path)

		if err != nil {
			return err
		}
	}

	for {
		select {
		case ev := <-watcher.Events:
			// console.Debug("change %d %s", ev.Op, ev.Name)
			if ev.Op&fsnotify.Write == fsnotify.Write || ev.Op&fsnotify.Create == fsnotify.Create {
				for _, v := range self.fileWatcher {
					if strings.Compare(v.Path, ev.Name) == 0 {
						if ok, err := v.CompareHash(); err != nil {
							console.Err("goflow: reload config compare hash failed! %s", err)

							break
						} else if !ok {
							console.Log("goflow: reload application notify %s", ev.Name)
							// reload all config
							if err := self.ReloadApplication(); err == cfg.ErrNeedShutdown {
								self.Shutdown()
							}
						}

						break
					}
				}
			}
		case err := <-watcher.Errors:
			return err
		}
	}
}

func (self *WebApplication) Loop() error {
	console.Ok("goflow: web app %s running on %s %s", self.Config.App.Name, self.Config.Listen.Type, self.Config.Listen.Addr)

	err := fasthttp.Serve(self.Listener, self.handle)

	if err != nil {
		// console.Err("fasthttp: %s", err)
		return err
	}

	return nil
}

func (self *WebApplication) Shutdown() {
	// console.Debug("goflow: force abort `%s` web application", self.Config.App.Name)
	if self.Listener != nil {
		self.Listener.Close()
		self.Listener = nil
	}

	if self.watcher != nil {
		self.watcher.Close()
		self.watcher = nil
	}
}

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

func (self *WebApplication) ReloadApplication() error {
	cfg, ctx, err := cfg.ReloadConfig(self.Context, self.Path)

	if err != nil {
		return err
	}

	if self.Config != nil {
		if cfg.Listen.Addr != self.Config.Listen.Addr || cfg.Listen.Type != self.Config.Listen.Type {
			// relisten
			self.Shutdown()
		}
	}

	if self.Listener == nil {
		ln, err := getListener(cfg.Listen.Type, cfg.Listen.Addr)

		if err != nil {
			return err
		}

		self.Listener = ln
	}

	self.Context = ctx
	self.Config = cfg

	return nil
}

func NewWebApplication(path string) (*WebApplication, error) {
	app := new(WebApplication)
	app.Path = path
	err := app.ReloadApplication()

	if err != nil {
		return nil, err
	}

	return app, nil
}
