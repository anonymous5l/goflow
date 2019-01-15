package hello_world

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

	ctx := req.GetContext()
	ctx.SetBody([]byte("Hello World!"))

	return nil
}

func LogicAfter(req interfaces.Request) error {

	console.Log("After request")

	return nil
}

func FlowInit(ctx interfaces.Context, params interface{}) error {
	console.Ok("mount hello world")

	// all request through `Logic`
	ctx.BeforeMiddleware(LogicBefore)

	ctx.Register("/", general.GET, Logic)

	// all request end `LogicAfter`
	ctx.AfterMiddleware(LogicAfter)

	return nil
}
