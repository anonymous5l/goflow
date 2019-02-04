package cfg

import (
	"errors"
	"strings"
)

type PluginStatus int

func (s PluginStatus) String() string {
	var array []string

	if s == Uninitialized {
		array = append(array, "Uninitialized")
	}

	if s&Initialization == Initialization {
		array = append(array, "Initialization")
	}

	if s&Reloading == Reloading {
		array = append(array, "Reloading")
	}

	if s&Complated == Complated {
		array = append(array, "Complated")
	}

	return strings.Join(array, "|")
}

const (
	Uninitialized  PluginStatus = 0
	Initialization PluginStatus = 1
	Reloading      PluginStatus = 1 << 1
	Complated      PluginStatus = 1 << 2
	Damaged        PluginStatus = 1 << 3
)

const (
	InitFuncName   = "Init"
	UninitFuncName = "Uninit"
)

var (
	ErrOnInit = errors.New("on plugin init")
)
