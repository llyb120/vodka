package plugin

import (
	"strings"
	"sync"
	"vodka/xml"
)

var tagHandlers = sync.Map{}

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
