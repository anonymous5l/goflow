package impl

import (
	"errors"
	"github.com/anonymous5l/console"
	"github.com/anonymous5l/goflow/general"
	"github.com/anonymous5l/goflow/interfaces"
	"github.com/valyala/fasthttp"
	"plugin"
	"reflect"
	"sync"
)

type FlowContext struct {
	handle func(interfaces.Request) error
}

type ServiceContext struct {
	p *plugin.Plugin
}

type ContextImpl struct {
	scope    []*ScopeImpl
	services map[string]*ServiceContext
	env      map[string]interface{}
	mutex    *sync.RWMutex
}

// errors

var ErrServiceAlreadyExists = errors.New("services already exists")
var ErrServiceNull = errors.New("services can't be null")
var ErrServiceNotFound = errors.New("services not found")
var ErrExecute = errors.New("invoke failed!")

func NewContextImpl(env map[string]interface{}) *ContextImpl {
	s := make(map[string]*ServiceContext)

	return &ContextImpl{
		services: s,
		env:      env,
		mutex:    &sync.RWMutex{},
	}
}

func (ctx *ContextImpl) SwitchEnv(env map[string]interface{}) {
	ctx.mutex.Lock()
	ctx.env = env
	ctx.mutex.Unlock()
}

func (ctx *ContextImpl) RegisterService(name string, p *plugin.Plugin) error {
	if p == nil {
		return ErrServiceNull
	}

	ctx.mutex.Lock()

	if _, ok := ctx.services[name]; ok {
		ctx.mutex.Unlock()
		return ErrServiceAlreadyExists
	}

	ctx.services[name] = &ServiceContext{
		p: p,
	}

	ctx.mutex.Unlock()
	return nil
}

func (ctx *ContextImpl) UnregisterService(name string) error {
	ctx.mutex.Lock()

	if _, ok := ctx.services[name]; ok {
		ctx.mutex.Unlock()
		return ErrServiceNotFound
	}

	delete(ctx.services, name)

	ctx.mutex.Unlock()

	return nil
}

func (ctx *ContextImpl) IterServices(f func(string, *ServiceContext)) {
	ctx.mutex.RLock()

	for k, v := range ctx.services {
		f(k, v)
	}

	ctx.mutex.RUnlock()
}

func (ctx *ContextImpl) Invoke(name string, method string, args ...interface{}) (res []interface{}, err error) {
	err = ErrServiceNotFound

	defer func() {
		if e := recover(); e != nil {
			var ok bool

			err, ok = e.(error)

			if !ok {
				err = ErrExecute
			}

			console.Err("goflow: invoke service `%s` method `%s` exception %s", name, method, e)
		}
	}()

	var member interface{}

	if p, ok := ctx.services[name]; ok {
		if f, err := p.p.Lookup(method); err != nil {
			return nil, err
		} else {
			member = f
		}

		res, err = member.(func(...interface{}) ([]interface{}, error))(args...)
		return res, err
	}

	return nil, err
}

func (ctx *ContextImpl) RefMember(name string, m string) (interface{}, error) {
	var member interface{}

	if p, ok := ctx.services[name]; ok {
		if f, err := p.p.Lookup(m); err != nil {
			return nil, err
		} else {
			member = f
		}

		return member, nil
	}

	return nil, ErrServiceNotFound
}

func (ctx *ContextImpl) Member(name string, m string) (interface{}, error) {
	if i, err := ctx.RefMember(name, m); err != nil {
		return nil, err
	} else {
		return eindirect(reflect.ValueOf(i)).Interface(), err
	}
}

func (ctx *ContextImpl) CompareMember(member interface{}, service string, name string) bool {
	if i, err := ctx.Member(service, name); err != nil {
		console.Err("goflow: CompareMember exception %s", err)
		return false
	} else {
		m := eindirect(reflect.ValueOf(member)).Interface()

		if i != m {
			return false
		}
	}

	return true
}

func (ctx *ContextImpl) GetEnv(key string) (interface{}, bool) {
	ctx.mutex.RLock()
	v, ok := ctx.env[key]
	ctx.mutex.RUnlock()
	return v, ok
}

func (ctx *ContextImpl) GetMapEnv(key string) (map[string]interface{}, bool) {
	if v, ok := ctx.GetEnv(key); ok {
		i, ok := v.(map[string]interface{})
		return i, ok
	}

	return nil, false
}

func (ctx *ContextImpl) findScope(scope *ScopeImpl) (int, *ScopeImpl) {
	for i, v := range ctx.scope {
		if v == scope {
			return i, v
		}
	}

	return -1, nil
}

func (ctx *ContextImpl) RegisterScope(scope *ScopeImpl) {
	ctx.mutex.Lock()
	if i, _ := ctx.findScope(scope); i == -1 {
		ctx.scope = append(ctx.scope, scope)
	}
	ctx.mutex.Unlock()
}

func (ctx *ContextImpl) UnregisterScope(scope *ScopeImpl) {
	// console.Log("unregister scope")

	ctx.mutex.Lock()
	if i, v := ctx.findScope(scope); v != nil {
		// console.Log("dispose scope %d %p", i, v)
		v.Dispose()

		ctx.scope = append(ctx.scope[:i], ctx.scope[i+1:]...)
	}
	ctx.mutex.Unlock()
}

func (ctx *ContextImpl) Handle(handle *fasthttp.RequestCtx) error {
	ctx.mutex.RLock()

	p := string(handle.Path())
	m := string(handle.Method())

	r := NewRequestImpl(handle)

	console.Log("goflow: %s %s", m, p)

	for _, s := range ctx.scope {
		err := s.Handle(handle, r, m, p)

		if err == general.Abort {
			break
		}
	}

	ctx.mutex.RUnlock()

	return nil
}

func eindirect(v reflect.Value) reflect.Value {
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		return eindirect(v.Elem())
	default:
		return v
	}
}
