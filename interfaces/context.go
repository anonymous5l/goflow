package interfaces

type Context interface {
	Invoke(string, string, ...interface{}) ([]interface{}, error)

	Member(string, string) (interface{}, error)

	CompareMember(interface{}, string, string) bool

	GetEnv(string) (interface{}, bool)
	GetMapEnv(string) (map[string]interface{}, bool)
}
