package plugin

import (
	"strings"
	"sync"
	"vodka/xml"
)

type NodeHandler func(builder *strings.Builder, node *xml.Node, params map[string]interface{}, resultParams *[]interface{}, root *xml.Node)

var nodeHandlers = sync.Map{}

func RegisterNode(name string, handler func(builder *strings.Builder, node *xml.Node, params map[string]interface{}, resultParams *[]interface{}, root *xml.Node)) {
	nodeHandlers.Store(name, handler)
}

func GetNodeHandler(name string) (NodeHandler, bool) {
	handler, ok := nodeHandlers.Load(name)
	if !ok {
		return nil, false
	}
	return handler.(NodeHandler), true
}

type HookType int

const (
	HOOK_AROUND_EXECUTE HookType = iota
)

type HookContext struct {
	Builder        **strings.Builder
	RequestParams  *[]interface{}
	ResultWrappers *[]interface{}
	Fn             func()
	// node *xml.Node
	// params map[string]interface{}
	// resultParams *[]interface{}
	// root *xml.Node
}

func RegisterHook(hookType HookType, handler func(context *HookContext)) {
	
}
