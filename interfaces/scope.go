package interfaces

import (
	"github.com/anonymous5l/goflow/general"
)

type Scope interface {
	Register(general.ContextHandleMethod, string, func(Request) error) error

	Before(func(Request) error) error
	After(func(Request) error) error
}
