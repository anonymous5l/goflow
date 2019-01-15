package interfaces

import (
	"github.com/anonymous5l/goflow/general"
)

type Context interface {
	Register(general.ContextHandleMethod, string, func(Request) error) error

	BeforeMiddleware(func(Request) error) error
	AfterMiddleware(func(Request) error) error

	Invoke(string, string, ...interface{}) ([]interface{}, error)

	Member(string, string) (interface{}, error)

	CompareMember(interface{}, string, string) bool

	GetEnv(string) (interface{}, bool)
	GetMapEnv(string) (map[string]interface{}, bool)
}
