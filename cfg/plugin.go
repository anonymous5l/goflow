package cfg

import (
	"github.com/anonymous5l/console"
	"github.com/anonymous5l/goflow/impl"
	"github.com/anonymous5l/goflow/interfaces"
	"github.com/anonymous5l/goflow/utils"
	"io"
	"os"
	"plugin"
)

type config_plugin struct {
	Name  string      `toml:"name"`
	Path  string      `toml:"path"`
	Extra interface{} `toml:"extra"`

	status PluginStatus   `toml:"-"`
	fh     *HashFile      `toml:"-"`
	p      *plugin.Plugin `toml:"-"`
}

type config_service struct {
	config_plugin
}

type config_flow struct {
	config_plugin

	scope *impl.ScopeImpl
}

func copyFile(dstName, srcName string) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		return
	}
	defer src.Close()
	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer dst.Close()
	return io.Copy(dst, src)
}

func (self *config_plugin) LoadSymbol() error {
	if self.fh == nil {
		if fh, err := NewHashFile(self.Path); err != nil {
			return err
		} else {
			self.fh = fh
		}
	}

	p, err := plugin.Open(self.Path)

	if err != nil {
		return err
	}

	// console.Debug("change plugin %p %p", self.p, p)

	self.p = p

	return nil
}

func (self *config_plugin) recoverError(err *error) {
	if e := recover(); e != nil {
		if ec, ok := e.(error); ok {
			*err = ec
		} else {
			*err = ErrOnInit
		}

		console.Err("goflow: plugin recover exception %s %s", self.Name, utils.ErrorStack(6, *err))
	}
}

func (self *config_plugin) lookupFunc(fname string) (f interface{}, err error) {
	f, err = self.p.Lookup(fname)

	return f, err
}

func (self *config_service) Init(ctx *impl.ContextImpl) (err error) {
	defer self.recoverError(&err)

	if f, err := self.lookupFunc(InitFuncName); err == nil {
		if err := f.(func(interfaces.Context, interface{}) error)(ctx, self.Extra); err != nil {
			return err
		} else if err := ctx.RegisterService(self.Name, self.p); err != nil {
			return err
		}
	}

	return err
}

func (self *config_service) Uninit(ctx *impl.ContextImpl) (err error) {
	defer self.recoverError(&err)

	if f, err := self.lookupFunc(UninitFuncName); err == nil {
		if err := f.(func(interfaces.Context, interface{}) error)(ctx, self.Extra); err != nil {
			return err
		}
	}

	ctx.UnregisterService(self.Name)

	return err
}

func (self *config_flow) Init(ctx *impl.ContextImpl) (err error) {
	defer self.recoverError(&err)

	if f, err := self.lookupFunc(InitFuncName); err == nil {
		if self.scope == nil || self.scope.Disposed {
			self.scope = impl.NewScopeImpl(ctx)
		}

		if err := f.(func(interfaces.Context, interfaces.Scope, interface{}) error)(ctx, self.scope, self.Extra); err != nil {
			return err
		}
	}

	ctx.RegisterScope(self.scope)

	return err
}

func (self *config_flow) Uninit(ctx *impl.ContextImpl) (err error) {
	defer self.recoverError(&err)

	if f, err := self.lookupFunc(UninitFuncName); err == nil {
		if err := f.(func(interfaces.Context, interface{}) error)(ctx, self.Extra); err != nil {
			return err
		}
	}

	ctx.UnregisterScope(self.scope)

	return err
}
