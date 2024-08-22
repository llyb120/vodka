package mapper

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	analyzer "github.com/llyb120/vodka/analyzer"
	database "github.com/llyb120/vodka/database"
	"github.com/llyb120/vodka/plugin"
)

var mappers map[string]*Mapper

//= make(map[string]*Mapper)
// var mappersLock = sync.RWMutex{}

// 缓存，不用每次都重新绑定
var mapperCache map[string]interface{}

// = make(map[string]interface{})
var mapperCacheLock = sync.RWMutex{}

// initmapper的锁
var mutex = sync.Mutex{}

type Mapper struct {
	FunctionMap map[string]*analyzer.Function
	// MapperItemsMap map[string]*MapperItem
	NameSpace string
}

type MapperItem struct {
	// Analyzer *analyzer.Analyzer
	Function *analyzer.Function
	MethodId string
}

func ScanEmbedMapper(staticFiles embed.FS) error {
	plugin.RegisterCommonFunc()
	var wg sync.WaitGroup
	var analyzers []*analyzer.Analyzer
	rwMutex := sync.RWMutex{}
	// 遍历嵌入的文件系统
	err := fs.WalkDir(staticFiles, "static", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 检查文件是否为XML
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".xml") {
			// 读取XML文件内容
			content, err := staticFiles.ReadFile(path)
			if err != nil {
				return err
			}
			wg.Add(1)
			go func(path string) {
				// 输出找到的XML文件路径
				fmt.Println("找到XML文件:", path)
				defer wg.Done()
				parser := analyzer.NewAnalyzer(string(content))
				parser.Parse()
				rwMutex.Lock()
				defer rwMutex.Unlock()
				analyzers = append(analyzers, parser)
			}(path)
		}
		return nil
	})
	wg.Wait()
	if err != nil {
		return err
	}
	// 整理所有的analyzer，将相同命名空间的mapper集合到一起
	return InitMappers(analyzers)
}

func ScanMapper(dir string) error {
	plugin.RegisterCommonFunc()
	var wg sync.WaitGroup
	var analyzers []*analyzer.Analyzer
	rwMutex := sync.RWMutex{}

	// 遍历指定目录
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 检查文件是否为XML
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".xml") {
			// 读取XML文件内容
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			wg.Add(1)
			go func(path string) {
				// 输出找到的XML文件路径
				println("找到XML文件:", path)
				defer wg.Done()
				parser := analyzer.NewAnalyzer(string(content))
				parser.Parse()
				rwMutex.Lock()
				defer rwMutex.Unlock()
				analyzers = append(analyzers, parser)
			}(path)
		}
		return nil
	})
	wg.Wait()
	// 关闭analyzersChan，释放
	if err != nil {
		return err
	}
	// 整理所有的analyzer，将相同命名空间的mapper集合到一起
	return InitMappers(analyzers)
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
		mapper = newMapper(analyzer.Namespace)
		mappers[analyzer.Namespace] = mapper
	}
	for _, function := range analyzer.Functions {
		mapper.FunctionMap[function.Id] = function
	}
}

// func GetMapper(namespace string) *Mapper {
// 	mutex.Lock()
// 	defer mutex.Unlock()
// 	return Mappers[namespace]
// }

func InitMapper(source interface{}) error {
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

	// 需要重新初始化的情况
	mapperCacheLock.Lock()
	defer mapperCacheLock.Unlock()
	var mapper *Mapper
	if mapper, ok = mappers[namespace]; !ok {
		// 双重检查
		mapper = newMapper(namespace)
		mappers[namespace] = mapper
		// return errors.New("InitMapper: 无法找到命名空间 " + namespace)
	}
	err := bindMapper(source, mapper, mapperValue, mapperType, v)
	if err != nil {
		return err
	}

	// 保存缓存，使用双检
	if _, ok := mapperCache[namespace]; !ok {
		mapperCache[namespace] = source
	}
	return nil
}

// 为其生成方法
func bindMapper(source interface{}, mapper *Mapper, mapperValue reflect.Value, mapperType reflect.Type, v reflect.Value) error {
	// 获得metadata
	metaData := NewMetaData(mapper.NameSpace, mapperValue)
	if metaData != nil {
		for _, function := range metaData.Functions {
			mapper.FunctionMap[function.Id] = function
		}
	}
	// 如果有 BuildTags 方法，直接调用
	// var customSqlMap map[string]string
	// customSqlMap := make(map[string]string)

	for i := 0; i < mapperType.NumField(); i++ {
		field := mapperType.Field(i)

		if field.Name == "VodkaMapper" {
			// 循环父类的字段
			vodkaMapperType := field.Type
			for j := 0; j < vodkaMapperType.NumField(); j++ {
				vodkaField := vodkaMapperType.Field(j)
				err := generateFunctionBody(mapper, v.Field(i), vodkaField, metaData)
				if err != nil {
					return err
				}
			}
			// log.Println("ok")
		} else {
			err := generateFunctionBody(mapper, v, field, nil)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func generateFunctionBody(mapper *Mapper, v reflect.Value, field reflect.StructField, metadata *MetaData) error {
	fieldType := field.Type

	// 如果是一个方法
	if field.Type.Kind() != reflect.Func {
		return nil
	}

	// 如果已经有实现了，则continue
	if !v.FieldByName(field.Name).IsNil() {
		return nil
	}

	paramNames := getParamNames(field)

	// 找到目标方法
	// 优先使用 XML 标签，如果没有则使用字段名
	methodName := field.Tag.Get("xml")
	if methodName == "" {
		methodName = field.Name
	}
	if _, ok := mapper.FunctionMap[methodName]; !ok {
		var sqlTag string
		sqlTag = field.Tag.Get("sql")

		// 检查是否有sql的tag
		if sqlTag != "" {
			// 生成一个sql的函数
			function, err := analyzer.ParseSingleSql(mapper.NameSpace, sqlTag)
			if err != nil {
				return err
			}
			mapper.FunctionMap[methodName] = function
		}
		// 暂时先不返回错误
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
				// 如果返回的结构体是interface{}，且metadata不为空（使用basemapper)
				if resultType.Elem().Kind() == reflect.Interface && metadata != nil {
					// 使用mockslice
					mockSlice := &database.MockSlice{
						Data: &[]interface{}{},
						Type: metadata.ModelType,
					}
					resultWrappers = append(resultWrappers, mockSlice)
				} else {
					result = reflect.New(resultType).Interface()
					resultWrappers = append(resultWrappers, result)
				}
				// returns = append(returns, reflect.ValueOf(result))
			} else if resultType == reflect.TypeOf((*error)(nil)).Elem() {
				errIndexes = append(errIndexes, i)
				resultWrappers = append(resultWrappers, nil)
				// if err != nil {
				// 	returns = append(returns, reflect.ValueOf(err))
				// } else {
				// 	returns = append(returns, reflect.Zero(resultType))
				// }
			} else if resultType == reflect.TypeOf((*int64)(nil)).Elem() {
				result = new(int64)
				resultWrappers = append(resultWrappers, result)
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

		err := analyzer.CallFunction(mapper.FunctionMap[methodName], params, resultWrappers)
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

func newMapper(namespace string) *Mapper {
	return &Mapper{
		// MapperItemsMap: make(map[string]*MapperItem),
		FunctionMap: make(map[string]*analyzer.Function),
		NameSpace:   namespace,
	}
}
