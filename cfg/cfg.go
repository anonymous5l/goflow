package cfg

import (
	"errors"
	"github.com/BurntSushi/toml"
	"github.com/anonymous5l/console"
	"github.com/anonymous5l/goflow/impl"
	"github.com/anonymous5l/goflow/interfaces"
	"github.com/anonymous5l/goflow/utils"
	"plugin"
)

type config_listen struct {
	Type string `toml:"type"`
	Addr string `toml:"addr"`
}

type config_plugin struct {
	Name  string      `toml:"name"`
	Path  string      `toml:"path"`
	Extra interface{} `toml:"extra"`

	p *plugin.Plugin `toml:"-"`
}

var ErrOnInit = errors.New("on plugin init")

func (self *config_plugin) LoadSymbol() error {
	p, err := plugin.Open(self.Path)

	if err != nil {
		return err
	}

	self.p = p

	return nil
}

type config_service struct {
	config_plugin
}

func (self *config_service) Init(ctx interfaces.Context) (err error) {
	defer func() {
		if e := recover(); e != nil {
			if ec, ok := e.(error); ok {
				err = ec
			} else {
				err = ErrOnInit
			}

			console.Err("goflow: service init exception %s", utils.ErrorStack(6, err))
		}
	}()

	f, err := self.p.Lookup("ServiceInit")

	if err != nil {
		return err
	}

	err = f.(func(interfaces.Context, interface{}) error)(ctx, self.Extra)

	if err != nil {
		return err
	}

	return nil
}

type config_flow struct {
	config_plugin
}

func (self *config_flow) Init(ctx interfaces.Context) (err error) {
	defer func() {
		if e := recover(); e != nil {
			if ec, ok := e.(error); ok {
				err = ec
			} else {
				err = ErrOnInit
			}

			console.Err("goflow: flow init exception %s", utils.ErrorStack(6, err))
		}
	}()

	f, err := self.p.Lookup("FlowInit")

	if err != nil {
		return err
	}

	err = f.(func(interfaces.Context, interface{}) error)(ctx, self.Extra)

	if err != nil {
		return err
	}

	return nil
}

type config struct {
	Listen      *config_listen         `toml:"listen"`
	Services    []*config_service      `toml:"service"`
	Flow        []*config_flow         `toml:"flow"`
	Environment map[string]interface{} `toml:"environment"`
}

var Listen *config_listen
var Services []*config_service
var Flow []*config_flow
var Environment map[string]interface{}

func InitConfig(path string) (*impl.ContextImpl, error) {
	var cfg *config

	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, err
	}

	ctx := impl.NewContextImpl(cfg.Environment)

	Listen = cfg.Listen

	for _, v := range cfg.Services {
		if err := v.LoadSymbol(); err != nil {
			console.Err("goflow: can't load `%s` service!", v.Name)
		} else {
			if err := v.Init(ctx); err == nil {
				// register to context
				if err := ctx.RegisterService(v.Name, v.p); err != nil {
					console.Err("goflow: already exists service `%s`", v.Name)
				}
			} else {
				console.Err("goflow: init service `%s` unexcept exception: %s", v.Name, err)
			}
		}
	}

	Services = cfg.Services

	for _, v := range cfg.Flow {
		if err := v.LoadSymbol(); err != nil {
			console.Err("goflow: can't load `%s` flow! %s", v.Name, err)
		} else {
			if err := v.Init(ctx); err != nil {
				console.Err("goflow: init flow `%s` unexcept exception: %s", v.Name, err)
			}
		}
	}

	Flow = cfg.Flow

	Environment = cfg.Environment

	return ctx, nil
}
