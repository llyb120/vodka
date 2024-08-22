package plugin

import (
	"github.com/llyb120/vodka/xml"
	"strings"
	"sync"
)

var tagHandlers = sync.Map{}
var functionHandlers = sync.Map{}

func GetTagHandler(name string) (CustomTagHandler, bool) {
	handler, ok := tagHandlers.Load(name)
	if !ok {
		return nil, false
	}
	return handler.(CustomTagHandler), true
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

type CustomTagHandler func(builder *strings.Builder, node *xml.Node, params map[string]interface{}, resultParams *[]interface{}, root *xml.Node)

func RegisterTag(tag string, handler CustomTagHandler) {
	tagHandlers.Store(strings.ToUpper(tag), handler)
}

func RegisterFunction(name string, handler func(args []interface{}) interface{}) {
	functionHandlers.Store(strings.ToUpper(name), handler)
}

func GetFunctionHandler(name string) (func(args []interface{}) interface{}, bool) {
	handler, ok := functionHandlers.Load(strings.ToUpper(name))
	if !ok {
		return nil, false
	}
	return handler.(func(args []interface{}) interface{}), true
}
