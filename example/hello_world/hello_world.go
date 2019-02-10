package main

import (
	"fmt"
	"github.com/anonymous5l/console"
	"github.com/anonymous5l/goflow/general"
	"github.com/anonymous5l/goflow/interfaces"
)

var context interfaces.Context

func LogicBefore(req interfaces.Request) error {

	console.Log("Before request")

	return nil
}

func Logic(req interfaces.Request) error {

	console.Log("Hello World")

	// make panic test response
	// a := []int{1, 2}
	// console.Log("make panic %d", a[3])

	// first arg config service name
	// second arg service func name
	// last func args
	ret, err := context.Invoke("hello_service", "ExposeFunc", 1, 2)

	if err != nil {
		console.Err("exec service error")
		return general.Abort
	}

	ctx := req.GetContext()
	ctx.SetBody([]byte(fmt.Sprintf("Hello World!\n%+v", ret)))

	return nil
}

func LogicAfter(req interfaces.Request) error {

	console.Log("After request")

	return nil
}

func Uninit(ctx interfaces.Context, params interface{}) error {
	console.Debug("hello_world: uninit")

	return nil
}

func Init(ctx interfaces.Context, scope interfaces.Scope, params interface{}) error {
	console.Ok("mount hello world")

	context = ctx

	// all request through `Logic`
	scope.Before(LogicBefore)

	scope.Register(general.GET, "/", Logic)

	// all request end `LogicAfter`
	scope.After(LogicAfter)

	return nil
}
