package general

import (
	"errors"
)

type ContextHandleMethod string

var GET ContextHandleMethod = ContextHandleMethod("GET")
var POST ContextHandleMethod = ContextHandleMethod("POST")
var PUT ContextHandleMethod = ContextHandleMethod("PUT")
var DELETE ContextHandleMethod = ContextHandleMethod("DELETE")
var PATCH ContextHandleMethod = ContextHandleMethod("PATCH")
var OPTIONS ContextHandleMethod = ContextHandleMethod("OPTIONS")

var End = errors.New("flow end")
var Abort = errors.New("flow abort")
