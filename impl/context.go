package impl

import (
	"bytes"
	"errors"
	"github.com/anonymous5l/console"
	"github.com/valyala/fasthttp"
	"goflow/general"
	"goflow/interfaces"
	"goflow/utils"
	"plugin"
	"reflect"
)

type FlowContext struct {
	handle func(interfaces.Request) error
}

type ServiceContext struct {
	cache map[string]interface{}

	p *plugin.Plugin
}

type ContextImpl struct {
	amiddleware []*FlowContext
	bmiddleware []*FlowContext
	router      map[general.ContextHandleMethod]map[string][]*FlowContext
	services    map[string]*ServiceContext
	env         map[string]interface{}
}

// errors

var ErrServiceAlreadyExists = errors.New("services already exists")
var ErrServiceNull = errors.New("services can't be null")
var ErrServiceNotFound = errors.New("services not found")
var ErrExecute = errors.New("invoke failed!")
var ErrFlowInner = errors.New("flow exception")

func NewContextImpl(env map[string]interface{}) *ContextImpl {
	r := make(map[general.ContextHandleMethod]map[string][]*FlowContext)
	s := make(map[string]*ServiceContext)

	return &ContextImpl{
		router:   r,
		services: s,
		env:      env,
	}
}

func (ctx *ContextImpl) Register(method general.ContextHandleMethod, path string, handle func(interfaces.Request) error) error {
	_, ok := ctx.router[method]

	if !ok {
		ctx.router[method] = make(map[string][]*FlowContext)
	}

	arr, _ := ctx.router[method][path]

	ctx.router[method][path] = append(arr, &FlowContext{
		handle: handle,
	})

	return nil
}

func (ctx *ContextImpl) BeforeMiddleware(handle func(interfaces.Request) error) error {
	ctx.bmiddleware = append(ctx.bmiddleware, &FlowContext{
		handle: handle,
	})

	return nil
}

func (ctx *ContextImpl) AfterMiddleware(handle func(interfaces.Request) error) error {
	ctx.amiddleware = append(ctx.amiddleware, &FlowContext{
		handle: handle,
	})

	return nil
}

func (ctx *ContextImpl) RegisterService(name string, p *plugin.Plugin) error {

	if p == nil {
		return ErrServiceNull
	}

	if _, ok := ctx.services[name]; ok {
		return ErrServiceAlreadyExists
	}

	ctx.services[name] = &ServiceContext{
		cache: make(map[string]interface{}),
		p:     p,
	}

	return nil
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
		if f, ok := p.cache[method]; ok {
			member = f
		} else {
			if f, err := p.p.Lookup(method); err != nil {
				return nil, err
			} else {
				member = f
				p.cache[name] = member
			}
		}

		res, err = member.(func(...interface{}) ([]interface{}, error))(args...)
		return res, err
	}

	return nil, err
}

func (ctx *ContextImpl) RefMember(name string, m string) (interface{}, error) {
	var member interface{}

	if p, ok := ctx.services[name]; ok {
		if f, ok := p.cache[m]; ok {
			member = f
		} else {
			if f, err := p.p.Lookup(m); err != nil {
				return nil, err
			} else {
				p.cache[name] = f
				member = f
			}
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
	v, ok := ctx.env[key]
	return v, ok
}

func (ctx *ContextImpl) GetMapEnv(key string) (map[string]interface{}, bool) {
	if v, ok := ctx.GetEnv(key); ok {
		i, ok := v.(map[string]interface{})
		return i, ok
	}

	return nil, false
}

func (ctx *ContextImpl) Handle(handle *fasthttp.RequestCtx) (err error) {

	defer func() {
		if e := recover(); e != nil {
			console.Err("goflow: context handle exception %s", e)

			handle.Response.Reset()
			handle.SetStatusCode(fasthttp.StatusInternalServerError)

			var buffer bytes.Buffer
			buffer.WriteString("Internal Server Error")

			var ok bool

			err, ok = e.(error)

			if !ok {
				err = ErrFlowInner
			}

			// special env

			if d, ok := ctx.GetEnv("debug"); ok {
				if debug, ok := d.(bool); ok && debug {
					// print internal server error stack
					buffer.WriteString("\n=====================\n")

					// print error stack for plugin skip top stack
					buffer.WriteString(utils.ErrorStack(6, err))
				}
			}

			handle.SetBody(buffer.Bytes())
		}
	}()

	p := string(handle.Path())
	m := string(handle.Method())

	console.Log("goflow: %s %s", m, p)

	request := NewRequestImpl(handle)

	for _, m := range ctx.bmiddleware {
		if err = m.handle(request); err == general.End {
			break
		}
	}

	if d, ok := ctx.router[general.ContextHandleMethod(m)]; ok {
		if farry, ok := d[p]; ok {
			for _, f := range farry {
				err = f.handle(request)

				if err == general.End {
					break
				}
			}
		}
	}

	for _, m := range ctx.amiddleware {
		if err = m.handle(request); err == general.End {
			break
		}
	}

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
