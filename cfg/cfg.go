package cfg

import (
	"errors"
	"github.com/BurntSushi/toml"
	"github.com/anonymous5l/console"
	"github.com/anonymous5l/goflow/impl"
	"github.com/anonymous5l/goflow/utils"
	"reflect"
	"strings"
	"sync"
)

var mutex *sync.RWMutex

func init() {
	mutex = &sync.RWMutex{}
}

type config_listen struct {
	Type string `toml:"type"`
	Addr string `toml:"addr"`
}

type config_app struct {
	Name string `toml:"name"`
}

type Config struct {
	App         *config_app            `toml:"app"`
	Listen      *config_listen         `toml:"listen"`
	Services    []*config_service      `toml:"service"`
	Flow        []*config_flow         `toml:"flow"`
	Environment map[string]interface{} `toml:"environment"`
}

var mServices []*config_service
var mFlow []*config_flow

var ErrNeedShutdown = errors.New("shutdown restart")

func Unload(ctx *impl.ContextImpl, p *config_plugin, uninitFunc func(*impl.ContextImpl) error) error {
	if p.status&Complated == Complated || p.status&Reloading == Reloading {
		err := uninitFunc(ctx)

		if err != nil {
			console.Err("goflow: unload plugin failed! %s %s", p.Name, err)
			p.status = Damaged
			return err
		}
	}

	return nil
}

func Reload(ctx *impl.ContextImpl, plugin *config_plugin, uninitFunc func(*impl.ContextImpl) error, initFunc func(*impl.ContextImpl) error) error {
	if plugin.status == Uninitialized {
		if plugin.fh == nil {
			fh, err := NewHashFile(plugin.Path)

			if err != nil {
				console.Err("goflow: update hash failed! %s `%s`", plugin.Name, plugin.Path)
				plugin.status = Damaged
				return err
			}

			plugin.fh = fh
			plugin.status = Initialization
		}
	} else if plugin.status&Complated == Complated {

		if compare, err := plugin.fh.CompareHash(); err != nil {
			console.Err("goflow: compare hash failed! %s `%s`", plugin.Name, plugin.Path)
			plugin.status = Damaged
			return err
		} else if !compare {
			plugin.status = Reloading
		}
	} else {
		// console.Debug("goflow: skip %s cause status in %s", plugin.Name, plugin.status.String())
		return nil
	}

	if plugin.status&Reloading == Reloading {
		// console.Debug("reloading plugin %p", plugin)
		if err := Unload(ctx, plugin, uninitFunc); err != nil {
			return err
		}
		return ErrNeedShutdown
	}

	if plugin.status&Complated != Complated {
		if err := plugin.LoadSymbol(); err != nil {
			console.Err("goflow: can't load `%s` plugin! %s", plugin.Name, err)
			plugin.status = Damaged
			return err
		} else {
			if err := initFunc(ctx); err != nil {
				console.Err("goflow: init plugin `%s` unexcept exception: %s", plugin.Name, err)
				plugin.status = Damaged
				return err
			}

			// console.Debug("init plugin success %s %p", plugin.Name, plugin)
			plugin.status = Complated
		}
	}
	// } else {
	// 	// console.Debug("nothing to change")
	// }

	return nil
}

func CrossCompare(t interface{}, p interface{}) int {
	switch reflect.TypeOf(t).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(t)
		d := reflect.ValueOf(p).Elem()

		for i := 0; i < s.Len(); i++ {
			sname := s.Index(i).Elem().FieldByName("Path").String()
			dname := d.FieldByName("Path").String()

			if strings.Compare(sname, dname) == 0 {
				return i
			}
		}
	}

	return -1
}

func ReloadConfig(ctx *impl.ContextImpl, path string) (*Config, *impl.ContextImpl, error) {
	var cfg *Config

	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, nil, err
	}

	if cfg.App == nil {
		cfg.App = new(config_app)
		cfg.App.Name = utils.GetRandomString(8, utils.STR_RANDOM)
	}

	if ctx == nil {
		ctx = impl.NewContextImpl(cfg.Environment)
	} else {
		ctx.SwitchEnv(cfg.Environment)
	}

	mutex.Lock()

	defer mutex.Unlock()

	// dected change remove old service
	for _, v := range mServices {
		if s := CrossCompare(cfg.Services, v); s == -1 {
			Unload(ctx, &v.config_plugin, v.Uninit)
		}
	}

	// dected change add or modify
	for i, v := range cfg.Services {
		if s := CrossCompare(mServices, v); s >= 0 {
			oldService := mServices[s]
			oldService.Name = v.Name
			oldService.Extra = v.Extra

			cfg.Services[i] = oldService

			Reload(ctx, &oldService.config_plugin, oldService.Uninit, oldService.Init)

			if oldService.status&Reloading == Reloading {
				return nil, nil, ErrNeedShutdown
			}
		} else {
			Reload(ctx, &v.config_plugin, v.Uninit, v.Init)
		}
	}

	// dected change remove old flow
	for _, v := range mFlow {
		if s := CrossCompare(cfg.Flow, v); s == -1 {
			Unload(ctx, &v.config_plugin, v.Uninit)
		}
	}

	// dected change add or modify
	for i, v := range cfg.Flow {
		if s := CrossCompare(mFlow, v); s >= 0 {
			oldFlow := mFlow[s]
			oldFlow.Name = v.Name
			oldFlow.Extra = v.Extra

			cfg.Flow[i] = oldFlow

			Reload(ctx, &oldFlow.config_plugin, oldFlow.Uninit, oldFlow.Init)

			if oldFlow.status&Reloading == Reloading {
				return nil, nil, ErrNeedShutdown
			}
		} else {
			Reload(ctx, &v.config_plugin, v.Uninit, v.Init)
		}
	}

	mServices = cfg.Services
	mFlow = cfg.Flow

	return cfg, ctx, nil
}
