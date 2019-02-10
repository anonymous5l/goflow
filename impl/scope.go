package impl

import (
	"bytes"
	"errors"
	"github.com/anonymous5l/console"
	"github.com/anonymous5l/goflow/general"
	"github.com/anonymous5l/goflow/interfaces"
	"github.com/anonymous5l/goflow/utils"
	"github.com/valyala/fasthttp"
	"sync"
)

type ScopeImpl struct {
	context     *ContextImpl
	amiddleware []*FlowContext
	bmiddleware []*FlowContext
	router      map[general.ContextHandleMethod]map[string][]*FlowContext
	mutex       *sync.RWMutex
	Disposed    bool
}

var ErrFlowInner = errors.New("flow exception")

func NewScopeImpl(ctx *ContextImpl) *ScopeImpl {
	scope := new(ScopeImpl)
	scope.context = ctx
	scope.mutex = &sync.RWMutex{}
	scope.router = make(map[general.ContextHandleMethod]map[string][]*FlowContext)

	return scope
}

func (scope *ScopeImpl) Register(method general.ContextHandleMethod, path string, handle func(interfaces.Request) error) error {
	scope.mutex.Lock()

	_, ok := scope.router[method]

	if !ok {
		scope.router[method] = make(map[string][]*FlowContext)
	}

	arr, _ := scope.router[method][path]

	scope.router[method][path] = append(arr, &FlowContext{
		handle: handle,
	})

	scope.mutex.Unlock()

	return nil
}

func (scope *ScopeImpl) Before(handle func(interfaces.Request) error) error {
	scope.mutex.Lock()
	scope.bmiddleware = append(scope.bmiddleware, &FlowContext{
		handle: handle,
	})
	scope.mutex.Unlock()

	return nil
}

func (scope *ScopeImpl) After(handle func(interfaces.Request) error) error {
	scope.mutex.Lock()
	scope.amiddleware = append(scope.amiddleware, &FlowContext{
		handle: handle,
	})
	scope.mutex.Unlock()

	return nil
}

func (scope *ScopeImpl) Handle(handle *fasthttp.RequestCtx, request interfaces.Request, m string, p string) (err error) {
	defer func() {
		scope.mutex.RUnlock()
	}()

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

			if d, ok := scope.context.GetEnv("debug"); ok {
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

	scope.mutex.RLock()

	// request := NewRequestImpl(handle)

	for _, m := range scope.bmiddleware {
		err = m.handle(request)

		if err == general.End {
			break
		}

		if err == general.Abort {
			return err
		}
	}

	if d, ok := scope.router[general.ContextHandleMethod(m)]; ok {
		if farry, ok := d[p]; ok {
			for _, f := range farry {
				err = f.handle(request)

				if err == general.End {
					break
				}

				if err == general.Abort {
					return err
				}

				if err != nil {
					console.Err("goflow: %s %s flow error %s", m, p, err)
				}
			}
		}
	}

	for _, m := range scope.amiddleware {
		err = m.handle(request)

		if err == general.End {
			break
		}

		if err == general.Abort {
			return err
		}
	}

	return nil
}

func (scope *ScopeImpl) Dispose() {
	scope.mutex.Lock()
	scope.amiddleware = nil
	scope.bmiddleware = nil
	scope.router = nil
	scope.Disposed = true
	scope.mutex.Unlock()
}
