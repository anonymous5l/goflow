package impl

import (
	"errors"
	"sync"

	"github.com/anonymous5l/goflow/general"
	"github.com/anonymous5l/goflow/interfaces"
	"github.com/valyala/fasthttp"
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

func (scope *ScopeImpl) handleArray(flows []*FlowContext, handle *fasthttp.RequestCtx, request interfaces.Request, m string, p string) error {
	defer func() {
		scope.mutex.RUnlock()
	}()

	scope.mutex.RLock()

	for _, m := range flows {
		err := m.handle(request)

		if err == general.End {
			break
		}

		if err == general.Abort {
			return err
		}
	}

	return nil
}

func (scope *ScopeImpl) HandleBefore(handle *fasthttp.RequestCtx, request interfaces.Request, m string, p string) error {
	return scope.handleArray(scope.bmiddleware, handle, request, m, p)
}

func (scope *ScopeImpl) HandleAfter(handle *fasthttp.RequestCtx, request interfaces.Request, m string, p string) error {
	return scope.handleArray(scope.amiddleware, handle, request, m, p)
}

func (scope *ScopeImpl) Handle(handle *fasthttp.RequestCtx, request interfaces.Request, m string, p string) error {
	if d, ok := scope.router[general.ContextHandleMethod(m)]; ok {
		if farry, ok := d[p]; ok {
			if err := scope.handleArray(farry, handle, request, m, p); err != nil {
				return err
			}
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
