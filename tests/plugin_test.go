package tests

import (
	"fmt"
	"strings"
	"testing"
	"vodka/plugin"
	"vodka/xml"
)

func TestPlugin(t *testing.T) {
	t.Run("TestPlugin", func(t *testing.T) {
		plugin.RegisterTag("permission", func(builder *strings.Builder, node *xml.Node, params map[string]interface{}, resultParams *[]interface{}, root *xml.Node) {
			// 从属性获取key和value
			key := node.Attrs["key"]
			value := node.Attrs["value"]
			t.Log("TestPlugin")
			builder.WriteString(fmt.Sprintf(" and %s > %s", key, value))
		})
		Prepare(t)
		userMapper.TestCustomTag()
	})

	t.Run("TestFunction", func(t *testing.T) {
		plugin.RegisterFunction("sum", func(args []interface{}) interface{} {
			return args[0].(int64) + args[1].(int64) + args[2].(int64)
		})
		Prepare(t)
		userMapper.TestFunction()
	})
}
