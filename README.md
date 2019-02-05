Goflow
===========

## 关于

基于 fasthttp 与 go1.8+ 的插件特性搭建的Web服务器框架 

设计初衷是为了快速搭建接口服务器 并且灵活进行 每个接口的控制

example 目录下有插件开发例子

代码中保留了对插件的热更新支持 通过子进程的方式实现

因为 特性问题导致 go 无法对插件进行卸载 所以目前只能通过这种方式实现热更新 

## 实例

```bash
$ go build
$ cd example
$ make
$ ../goflow -c config.toml
```

## 使用
	goflow -c [config_file]

## 配置文件规则

基本配置教程

```toml
[app]
name = "web application name"

[listen]
type = "tcp"             # unix 文件描述符 或 tcp 链接
addr = "127.0.0.1:8080"  # unix 类型的话则为文件地址 例如 `/tmp/goflow.sock` tcp 类型的话则为 `:8080` 或 `0.0.0.0:8080`

[environment]            # 环境变量
debug = true             # 内建字段 默认为 `false` 控制当Flow执行异常 是否 回显错误堆栈到返回体
somethingelse = "..."    # `interfaces.Context`.GetEnv("somethingelse") -> interface{} -> string

[environment.obj]
userDefine = "..."       # `interfaces.Context`.GetEnv("obj") -> interface{} -> map[string]interface{}

[[service]]              # 表示该插件为一个服务插件
name = ""                # 插件名称全局唯一
path = ""                # 插件路径
[service.extra]          # 额外拓展参数 
xxx  = ""

# 配置文件的顺序很重要 因为一个借口可能会经过多个流程控制

[[flow]]                 # 表示该插件为一个流程插件
name = ""                # 插件名
path = ""                # 插件路径
[flow.extra]             # 额外拓展参数
xxx = ""

```

进阶配置教程 下面示例为自带的`plugins`目录下的`json_params_check`使用方法

```toml
[app]
name = "web application name"

[listen]
type = "tcp"                                             # unix 文件描述符 或 tcp 链接
addr = "127.0.0.1:8080"                                  # unix 类型的话则为文件地址 例如 `/tmp/goflow.sock` tcp 类型的话则为 `:8080` 或 `0.0.0.0:8080`

[environment]                                            # 环境变量
debug = true                                             # 内建字段 默认为 `false` 控制当Flow执行异常 是否 回显错误堆栈到返回体

[[flow]]                                                 # 表示该插件为一个流程插件
name = "jpc"                                             # 插件名
path = "plugins/json_params_check/json_params_check.so"  # 插件路径

[[flow.extra.nodes]]                                     # 为插件特定的参数
path = "/hello"                                          # 对 `/hello` 接口进行json参数过滤
[flow.extra.nodes.params.arg1]                           # 表示 检查 上传的 arg1 参数
passing = false                                          # 是否必传参数
type = "str"                                             # 参数类型包括 `str` `num` `map` `bool`
[flow.extra.nodes.params.arg2]
passing = false
type = "map"
[flow.extra.nodes.params.arg2.params.arg1]               # 对 `map` 参数进行检查
passing = false
type = "str"
[flow.extra.nodes.params.arg3]
passing = false
type = "str"

# 上面的参数检查插件通过后进入下面的 业务逻辑插件

[[flow]]
name = "logic"
path = "logic.so"

```

## 快速上手

针对于服务插件

```go

package main

import (
	"github.com/anonymous5l/goflow/interfaces"
)

// 对外接口原型
func ExposeFunc(args ...interface{}) ([]interface{}, error) {
	return args, nil
}

// 内建方法 当服务器热更新时或配置文件变动导致插件卸载时 会调用此方法
// 通常表示 服务卸载
func Uninit(ctx interfaces.Context, params interface{}) error {
	return nil
}

// 内建方法 当服务器加载插件时发生
// params 表示为拓展参数 与 配置文件的 `service.extra` 对应
func Init(ctx interfaces.Context, params interface{}) error {
	return nil
}

```

针对于流程业务逻辑的插件

```go

package main

import (
	"github.com/anonymous5l/goflow/general"
	"github.com/anonymous5l/goflow/interfaces"
)

var context interfaces.Context

func LogicBefore(req interfaces.Request) error {
	// 过滤器 所有请求前的接口都会通过于此 可做一些请求前的处理
	return nil
}

func Logic(req interfaces.Request) error {
	// 注册的业务逻辑

	ctx := req.GetContext()
	ctx.SetBody([]byte("Hello World!"))

	return nil
}

func LogicAfter(req interfaces.Request) error {
	// 过滤器 所有请求后的接口都会通过于此
	return nil
}

// 插件卸载时发生
func Uninit(ctx interfaces.Context, params interface{}) error {
	
}

// 插件加载时发生
func Init(ctx interfaces.Context, scope interfaces.Scope, params interface{}) error {
	context = ctx

	// all request through `Logic`
	scope.Before(LogicBefore)

	// 注册根请求到Logic
	scope.Register(general.GET, "/", Logic)

	// all request end `LogicAfter`
	scope.After(LogicAfter)

	return nil
}

```

[Go Doc](https://godoc.org/github.com/anonymous5l/goflow)