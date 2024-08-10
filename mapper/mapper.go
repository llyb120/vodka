package mapper

import (
	"errors"
	"reflect"
	"strings"
	"sync"
	analyzer "vodka/analyzer"
)

var mappers = make(map[string]*Mapper)

// 缓存，不用每次都重新绑定
var mapperCache = make(map[string]interface{})
var mapperCacheLock = sync.RWMutex{}

// initmapper的锁
var mutex = sync.Mutex{}

type Mapper struct {
	MapperItemsMap map[string]*MapperItem
}

type MapperItem struct {
	Analyzer *analyzer.Analyzer
	MethodId string
}

func InitMappers(analyzers []*analyzer.Analyzer) error {
	mutex.Lock()
	defer mutex.Unlock()
	// 初始化
	mappers = make(map[string]*Mapper)
	mapperCache = make(map[string]interface{})
	for _, analyzer := range analyzers {
		initMapper(analyzer)
	}
	return nil
}

// 外部无法调用，只能通过InitMappers调用，所以这里就不需要加锁了
func initMapper(analyzer *analyzer.Analyzer) {
	mapper, ok := mappers[analyzer.Namespace]
	if !ok {
		mapper = &Mapper{
			MapperItemsMap: make(map[string]*MapperItem),
		}
		mappers[analyzer.Namespace] = mapper
	}
	for _, function := range analyzer.Functions {
		mapper.MapperItemsMap[function.Id] = &MapperItem{
			Analyzer: analyzer,
			MethodId: function.Id,
		}
	}
}

// func GetMapper(namespace string) *Mapper {
// 	mutex.Lock()
// 	defer mutex.Unlock()
// 	return Mappers[namespace]
// }

func BindMapper(source interface{}) error {
	mapperValue := reflect.ValueOf(source)
	// 确保传入的是指针
	if mapperValue.Kind() != reflect.Ptr {
		return errors.New("InitMapper: 参数必须是指针")
	}
	v := mapperValue.Elem()
	mapperType := v.Type()
	// 查找命名空间，为每个方法生成一个函数
	namespace := mapperType.Name()
	// 优先查找缓存中是否有，如果有，直接设置指针内容（必定是相同的）
	mapperCacheLock.RLock()
	existMapper, ok := mapperCache[namespace]
	mapperCacheLock.RUnlock()
	if ok {
		v.Set(reflect.ValueOf(existMapper).Elem())
		return nil
	}
	mapper, ok := mappers[namespace]
	if !ok {
		return errors.New("InitMapper: 无法找到命名空间 " + namespace)
	}
	err := bindMapper(source, mapper, mapperValue, mapperType, v)
	if err != nil {
		return err
	}
	mapperCacheLock.Lock()
	defer mapperCacheLock.Unlock()
	mapperCache[namespace] = source
	return nil
}

// 为其生成方法
func bindMapper(source interface{}, mapper *Mapper, mapperValue reflect.Value, mapperType reflect.Type, v reflect.Value) error {
	for i := 0; i < mapperType.NumField(); i++ {
		field := mapperType.Field(i)
		fieldType := field.Type

		// 如果是一个方法
		if fieldType.Kind() != reflect.Func {
			continue
		}

		paramNames := getParamNames(field)

		// 找到目标方法
		mapperItem, ok := mapper.MapperItemsMap[field.Name]
		if !ok {
			continue
			// 暂时先不返回错误
			//return errors.New("BindMapper: 无法找到方法 " + field.Name)
		}

		// 创建函数
		fn := reflect.MakeFunc(fieldType, func(args []reflect.Value) (results []reflect.Value) {
			// 这里是函数体的实现
			// 整理参数
			params := make(map[string]interface{})
			for i, arg := range args {
				if paramNames != nil && len(paramNames) > i {
					params[paramNames[i]] = arg.Interface()
				}
			}

			// 准备结果容器
			// 根据返回值类型创建一个零值
			var result interface{}
			var resultWrappers []interface{}
			var returns []reflect.Value
			var errIndexes []int
			for i := 0; i < fieldType.NumOut(); i++ {
				resultType := fieldType.Out(i)
				// 如果需要返回一个列表
				if resultType.Kind() == reflect.Slice {
					result = reflect.New(resultType).Interface()
					resultWrappers = append(resultWrappers, result)
					// returns = append(returns, reflect.ValueOf(result))
				} else if resultType == reflect.TypeOf((*error)(nil)).Elem() {
					errIndexes = append(errIndexes, i)
					resultWrappers = append(resultWrappers, nil)
					// if err != nil {
					// 	returns = append(returns, reflect.ValueOf(err))
					// } else {
					// 	returns = append(returns, reflect.Zero(resultType))
					// }
				} else {
					// 如果是结构体的话，为了成功返回nil，这里必须产生一个指针的指针，即**Struct
					//result = reflect.New(resultType).Interface()
					// 必须先产生指针内容
					value := reflect.New(resultType.Elem())
					valuePtr := reflect.New(resultType)
					valuePtr.Elem().Set(value)
					result = valuePtr.Interface()
					resultWrappers = append(resultWrappers, result)
					// if result != nil {
					// 	returns = append(returns, reflect.ValueOf(result))
					// } else {
					// 	returns = append(returns, reflect.Zero(resultType))
					// }
				}
			}

			err := mapperItem.Analyzer.Call(field.Name, params, resultWrappers)
			if err != nil {
				// 指定的替换为错误
				for _, index := range errIndexes {
					resultWrappers[index] = err
				}
			}
			// 将resultWrappers转换为returns
			for i, wrapper := range resultWrappers {
				if wrapper == nil {
					// 如果是error类型且为nil,返回零值
					// if fieldType.Out(i).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
					returns = append(returns, reflect.Zero(fieldType.Out(i)))
					// } else {
					// returns = append(returns, reflect.Zero(fieldType.Out(i)))
					// }
				} else {
					// 对于其他类型,直接返回wrapper的值
					wrapperValue := reflect.ValueOf(wrapper)
					if fieldType.Out(i).Kind() == reflect.Interface && fieldType.Out(i).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
						// 对于error接口类型，直接使用原始的error值
						returns = append(returns, wrapperValue)
					} else if wrapperValue.Elem().Kind() == reflect.Slice {
						// 对于切片类型，直接返回
						returns = append(returns, wrapperValue.Elem())
					} else {
						// 对于其他类型，返回Elem()
						returns = append(returns, wrapperValue.Elem())
					}
					// returns = append(returns, reflect.ValueOf(wrapper).Elem())
				}
			}

			// 如果有错误,将错误设置为最后一个返回值
			// if err != nil {
			// 	returns[len(returns)-1] = reflect.ValueOf(err)
			// }
			return returns
		})

		v.FieldByName(field.Name).Set(fn)
	}

	return nil
}

// 获取参数名称
func getParamNames(field reflect.StructField) []string {
	tag := field.Tag.Get("params")
	if tag == "" {
		return nil
	}
	return strings.Split(tag, ",")
}
