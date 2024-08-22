package vodka

import (
	"vodka/mapper"
	"vodka/plugin"
)

func ScanMapper(dir string) error {
	return mapper.ScanMapper(dir)
}

func InitMapper(source interface{}) error {
	return mapper.InitMapper(source)
}

func init() {
	plugin.RegisterFunction("len", func(args []interface{}) interface{} {
		if len(args) != 1 {
			return 0
		}
		// 必定是slice
		slice, ok := args[0].([]interface{})
		if !ok {
			return 0
		}
		return len(slice)
	})
}
