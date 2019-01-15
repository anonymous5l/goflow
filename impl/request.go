package impl

import (
	"encoding/json"
	"github.com/anonymous5l/goflow/interfaces"
	"github.com/valyala/fasthttp"
)

type RequestImpl struct {
	ctx *fasthttp.RequestCtx

	m map[string]interface{}
}

func NewRequestImpl(ctx *fasthttp.RequestCtx) *RequestImpl {
	return &RequestImpl{
		ctx: ctx,
		m:   make(map[string]interface{}),
	}
}

func (self *RequestImpl) GetContext() *fasthttp.RequestCtx {
	return self.ctx
}

func (self *RequestImpl) SetValue(key string, value interface{}) {
	self.m[key] = value
}

func (self *RequestImpl) GetValue(key string) (interface{}, bool) {
	v, o := self.m[key]
	return v, o
}

func (self *RequestImpl) Body() []byte {
	return self.ctx.PostBody()
}

func (self *RequestImpl) Method() string {
	return string(self.ctx.Method())
}

func (self *RequestImpl) URI() *fasthttp.URI {
	return self.ctx.URI()
}

func (self *RequestImpl) QueryArgs() *fasthttp.Args {
	return self.ctx.QueryArgs()
}

func (self *RequestImpl) JsonBody() (interface{}, error) {
	var result interface{}

	err := json.Unmarshal(self.Body(), &result)

	if err != nil {

		return nil, err
	}

	return result, nil
}

func (self *RequestImpl) JsonMapBody() (map[string]interface{}, error) {
	if raw, err := self.JsonBody(); err != nil {
		return nil, err
	} else if args, ok := raw.(map[string]interface{}); ok {
		return args, nil
	} else {
		return nil, interfaces.ErrBodyCovert
	}
}

func (self *RequestImpl) JsonArrayBody() ([]interface{}, error) {
	if raw, err := self.JsonBody(); err != nil {
		return nil, err
	} else if args, ok := raw.([]interface{}); ok {
		return args, nil
	} else {
		return nil, interfaces.ErrBodyCovert
	}
}
