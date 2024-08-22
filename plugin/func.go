package plugin

import (
	"reflect"
	"sync"
)

var once sync.Once

func RegisterCommonFunc() {
	once.Do(func() {
		RegisterFunction("len", _len)
	})
}

func _len(args []interface{}) interface{} {
	if len(args) != 1 {
		return 0
	}
	// 反射验证必定是slice
	v := reflect.ValueOf(args[0])
	if v.Kind() != reflect.Slice {
		return 0
	}
	return v.Len()
}
