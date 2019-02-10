package main

import (
	"github.com/anonymous5l/console"
	"github.com/anonymous5l/goflow/interfaces"
)

// Fixed method prototype
func ExposeFunc(args ...interface{}) ([]interface{}, error) {
	console.Debug("args0: %v", args[0])
	console.Debug("args1: %v", args[1])

	return args, nil
}

func Uninit(ctx interfaces.Context, params interface{}) error {
	console.Debug("hello_services: uninit")

	return nil
}

func Init(ctx interfaces.Context, params interface{}) error {
	console.Debug("hello_services: hello!")

	extra := params.(map[string]interface{})
	configArg := extra["someExtraArg"].(string)

	console.Debug("someExtraArg: %s", configArg)

	return nil
}
