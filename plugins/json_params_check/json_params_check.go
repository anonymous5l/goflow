package main

import (
	"github.com/anonymous5l/console"
	// "github.com/valyala/fasthttp"
	"github.com/anonymous5l/goflow/general"
	"github.com/anonymous5l/goflow/interfaces"
	"github.com/mitchellh/mapstructure"
	"reflect"
	"strings"
)

type Rule struct {
	Passing  bool
	PType    string      `mapstructure:"type"`
	PDefault interface{} `mapstructure:"default"`
	Params   map[string]*Rule
}

type Rules struct {
	Path   string
	Params map[string]*Rule
}

var arules []*Rules

func setResponseCode(req interfaces.Request, code int) error {
	ctx := req.GetContext()
	ctx.SetStatusCode(code)
	return general.Abort
}

func checkPType(t string, k reflect.Kind) bool {
	switch t {
	case "str":
		if k == reflect.String {
			return true
		}
	case "map":
		if k == reflect.Map {
			return true
		}
	case "num":
		if k == reflect.Float64 {
			return true
		}
	case "bool":
		if k == reflect.Bool {
			return true
		}
	}

	return false
}

func checkParamsObject(params map[string]*Rule, m map[string]interface{}) bool {
	if params != nil {
		for k, v := range params {
			if cv, ok := m[k]; ok {
				kind := reflect.TypeOf(cv).Kind()
				if !checkPType(v.PType, kind) {
					console.Warn("jpc: param type wrong %s", k)
					return false
				}

				if kind == reflect.Map {
					if !checkParamsObject(v.Params, cv.(map[string]interface{})) {
						return false
					}
				}
			} else {
				if !v.Passing && v.PDefault == nil {
					console.Warn("jpc: missing param %s", k)
					return false
				} else {
					m[k] = v.PDefault
				}
			}
		}
	}

	return true
}

func before_params_check(req interfaces.Request) error {
	uri := req.URI()
	path := string(uri.Path())

	for _, rule := range arules {
		if strings.Compare("POST", req.Method()) == 0 {
			if strings.Compare(path, rule.Path) == 0 {
				// found it
				if m, err := req.JsonMapBody(); err != nil {
					console.Warn("jpc: check %s unmarshal json failed!", path)
					return setResponseCode(req, 400)
				} else {
					if !checkParamsObject(rule.Params, m) {
						console.Warn("jpc: check %s params failed!", path)
						return setResponseCode(req, 400)
					}

					req.SetValue("JsonBody", m)
				}
			}
		}
	}

	return nil
}

func Init(ctx interfaces.Context, scope interfaces.Scope, params interface{}) error {
	console.Ok("mount json params check plugin")

	if p, ok := params.(map[string]interface{}); ok {
		if nodes, ok := p["nodes"].([]map[string]interface{}); ok {
			if err := mapstructure.Decode(nodes, &arules); err != nil {
				console.Err("jpc: can't convert to struct %s", err)
			}
		}
	}

	scope.Before(before_params_check)

	return nil
}
