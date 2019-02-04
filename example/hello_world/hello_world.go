package main

import (
	"github.com/anonymous5l/console"
	"github.com/anonymous5l/goflow/general"
	"github.com/anonymous5l/goflow/interfaces"
)

func LogicBefore(req interfaces.Request) error {

	console.Log("Before request")

	return nil
}

func Logic(req interfaces.Request) error {

	console.Log("Hello World")

	// make panic test response
	// a := []int{1, 2}
	// console.Log("make panic %d", a[3])

	ctx := req.GetContext()
	ctx.SetBody([]byte("Hello World!"))

	return nil
}

func LogicAfter(req interfaces.Request) error {

	console.Log("After request")

	return nil
}

func Init(ctx interfaces.Context, scope interfaces.Scope, params interface{}) error {
	console.Ok("mount hello world")

	// all request through `Logic`
	scope.Before(LogicBefore)

	scope.Register(general.GET, "/", Logic)

	// all request end `LogicAfter`
	scope.After(LogicAfter)

	return nil
}
