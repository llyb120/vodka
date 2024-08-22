package analyzer

import (
	"errors"
	"fmt"
	database "github.com/llyb120/vodka/database"
	mysqld "github.com/llyb120/vodka/database/mysql"
	"github.com/llyb120/vodka/plugin"
	page "github.com/llyb120/vodka/plugin/page"
	runner "github.com/llyb120/vodka/runner"
	"github.com/llyb120/vodka/xml"
	"log"
	"reflect"
	"regexp"
	"runtime/debug"
	"strings"
)

const (
	Select = iota
	Update
	Insert
	Delete
)

type Function struct {
	Id     string                                                                  //方法名
	Type   string                                                                  //方法类型
	Mapper string                                                                  //所属的mapper
	Func   func(resultWrappers []interface{}, params map[string]interface{}) error //方法体
}

type Functions Function //map[string]func(params map[string]interface{}) (interface{}, error)

type Analyzer struct {
	// 公有字段
	Namespace string               // 名称
	Functions map[string]*Function // 函数列表

	// 私有字段
	xmlContent string
	inited     bool
}

// ---------------------- 以下为公有方法 ----------------------

func NewAnalyzer(xmlContent string) *Analyzer {
	analyzer := &Analyzer{
		Functions:  make(map[string]*Function),
		inited:     false,
		xmlContent: xmlContent,
	}
	return analyzer
}

func CallFunction(fn *Function, params map[string]interface{}, resultWrappers []interface{}) error {
	// 任何时候，如果params中只有一个参数，且为map或者结构体指针的情况下，将其展开放入params中
	for k, v := range params {
		// 如果k以...开头，则默认展开
		if strings.HasPrefix(k, "...") {
			// 将...的值赋值给k
			params[strings.TrimPrefix(k, "...")] = v
			delete(params, k)
			// 展开
			extractObject(v, params)
			// for mk, mv := range v.(map[string]interface{}) {
			// 	params[mk] = mv
			// }
		}
	}

	if len(params) == 1 {
		for _, v := range params {
			// params[k] = reflect.ValueOf(v).Interface()
			// 如果是map
			extractObject(v, params)
			// if reflect.TypeOf(v).Kind() == reflect.Map {
			// 	for mk, mv := range v.(map[string]interface{}) {
			// 		params[mk] = mv
			// 	}
			// } else if reflect.TypeOf(v).Kind() == reflect.Ptr && reflect.TypeOf(v).Elem().Kind() == reflect.Struct {
			// 	extractStructFields(reflect.ValueOf(v).Elem(), params)
			// } else if reflect.TypeOf(v).Kind() == reflect.Struct {
			// 	extractStructFields(reflect.ValueOf(v), params)
			// }
			break
		}
	}

	return fn.Func(resultWrappers, params)
}

func (t *Analyzer) Call(id string, params map[string]interface{}, resultWrappers []interface{}) error {
	if !t.inited {
		return fmt.Errorf("未初始化")
	}
	fn, ok := t.Functions[id]
	if !ok {
		return fmt.Errorf("函数 %s 不存在", id)
	}
	return CallFunction(fn, params, resultWrappers)
}

func (t *Analyzer) Parse() error {
	if t.inited {
		return errors.New("已经初始化过了")
	}

	// 使用xml.Decoder来解析XML
	// 因为自带的库无法正确处理属性中的大小于号，干脆自己写一个xml解析器来进行解析
	parser := xml.NewParser(t.xmlContent)
	root, err := parser.Parse()
	if err != nil {
		return err
	}

	// 整理根节点
	// 根节点必须有namespace属性
	namespace, ok := root.Attrs["namespace"]
	if !ok {
		return errors.New("根节点必须有namespace属性")
	}
	// 设置命名空间
	t.Namespace = namespace
	for _, node := range root.Children {
		// node的attributes里必须有id属性，否则不处理
		id, ok := node.Attrs["id"]
		if !ok {
			continue
		}
		t.Functions[id] = generateFunction(namespace, node, root)
	}

	t.inited = true
	return nil
}

// 编译多个自定义sql
func ParseXml(namespace, xmlContent string) ([]*Function, error) {
	parser := xml.NewParser(xmlContent)
	root, err := parser.Parse()
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	functions := make([]*Function, 0)
	for _, node := range root.Children {
		functions = append(functions, generateFunction(namespace, node, root))
	}
	return functions, nil
}

// 用于将来的自动装配通用sql语句
func ParseMultiSql(namespace string, sqls []string) ([]*Function, error) {
	var builder strings.Builder
	builder.WriteString("<sql>")
	for _, s := range sqls {
		s = strings.TrimSpace(s)
		firstWord := strings.ToUpper(strings.Split(s, " ")[0])
		s = "<" + firstWord + ">" + s + "</" + firstWord + ">"
		builder.WriteString(s)
	}
	builder.WriteString("</sql>")
	parser := xml.NewParser(builder.String())
	root, err := parser.Parse()
	if err != nil {
		return nil, err
	}
	functions := make([]*Function, 0)
	for _, node := range root.Children {
		functions = append(functions, generateFunction(namespace, node, root))
	}
	return functions, nil
}

// 编译单个sql
func ParseSingleSql(namespace, sql string) (*Function, error) {
	// 获得sql的第一个单词
	sql = strings.TrimSpace(sql)
	firstWord := strings.ToUpper(strings.Split(sql, " ")[0])
	// 拼装响应的头
	sql = "<" + firstWord + ">" + sql + "</" + firstWord + ">"
	parser := xml.NewParser(sql)
	root, err := parser.Parse()
	if err != nil {
		return nil, err
	}
	return generateFunction(namespace, root, root), nil

}

// ---------------------- 以下为私有方法 ----------------------

// 生成对应的方法体
func generateFunction(mapperName string, node *xml.Node, root *xml.Node) *Function {
	funcFunc := func(resultWrappers []interface{}, params map[string]interface{}) (resultErr error) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("调用栈:\n%s\n", debug.Stack())
				resultErr = errors.New("捕获到错误")
			}
		}()
		var builder strings.Builder
		var invokeParams []interface{}
		for _, child := range node.Children {
			HandleNode(&builder, child, params, &invokeParams, root)
		}
		log.Printf("【%s】【%s】 sql : %s %v", mapperName, node.Attrs["id"], builder.String(), invokeParams)
		// 插件系统
		// 构造HOOK CONTEXT
		// todo: 插件后续再开发
		// builderPtr := &builder
		// hookContext := plugin.HookContext{
		// 	Builder:        &builderPtr,
		// 	RequestParams:  &invokeParams,
		// 	ResultWrappers: &resultWrappers,
		// }
		// fmt.Print(hookContext)
		// 如果是查询语句
		if node.Name == "SELECT" {
			// hookContext.Fn = func() {
			// 如果是在分页查询的环境下
			pg := page.GetPageContext()
			if pg != nil {
				// 这里表示已经开启了分页的，但是因为无法获得具体的泛型，没办法转换
				// 先查询总页数
				resultErr = page.QueryPage(mysqld.GetDB(), builder.String(), invokeParams, resultWrappers, pg)
				// total, _ := page.SelectTotal(mysqld.GetDB(), builder.String(), invokeParams)
				// pg.TotalRows = total
				//resultErr = database.QueryStruct(mysqld.GetDB(), builder.String(), invokeParams, resultWrappers)
			} else {
				resultErr = database.QueryStruct(mysqld.GetDB(), builder.String(), invokeParams, resultWrappers)
			}
			// }
			// 执行插件
			// plugin.Execute(hookContext)
		} else if node.Name == "INSERT" || node.Name == "UPDATE" || node.Name == "DELETE" {
			// 只能执行这三个指令
			// hookContext.Fn = func() {
			resultErr = database.ExecuteInt64(mysqld.GetDB(), builder.String(), invokeParams, resultWrappers)
			// }
		}
		return resultErr
	}

	return &Function{
		Mapper: mapperName,
		Id:     node.Attrs["id"],
		Type:   node.Name,
		Func:   funcFunc,
	}
}

// 处理节点
func HandleNode(builder *strings.Builder, node *xml.Node, params map[string]interface{}, resultParams *[]interface{}, root *xml.Node) {
	if node.Type == xml.Text {
		handleText(builder, node, params, resultParams)
	} else {
		switch node.Name {
		case "IF":
			handleIfStatement(builder, node, params, resultParams, root)
		case "FOREACH":
			handleForeachStatement(builder, node, params, resultParams, root)
		case "WHERE":
			handleWhereStatement(builder, node, params, resultParams, root)
		case "SET":
			handleSetStatement(builder, node, params, resultParams, root)
		case "SQL":
			handleSqlStatement(builder, node, params, resultParams, root)
		case "INCLUDE":
			handleIncludeStatement(builder, node, params, resultParams, root)
		default:
			// 处理自定义节点
			handleCustomStatement(builder, node, params, resultParams, root)
		}
	}
}

// 处理if语句
func handleIfStatement(builder *strings.Builder, node *xml.Node, params map[string]interface{}, resultParams *[]interface{}, root *xml.Node) {
	// 获取test属性的值
	testExpr, ok := node.Attrs["test"]
	if !ok {
		panic("if语句缺少test属性")
	}

	// 解析并计算test表达式
	result := runner.EvaluateExpression(testExpr, params)

	// 如果表达式结果为true，则处理if语句的子节点
	if result != false {
		for _, child := range node.Children {
			HandleNode(builder, child, params, resultParams, root)
		}
	}
}

// 处理foreach语句
func handleForeachStatement(builder *strings.Builder, node *xml.Node, params map[string]interface{}, resultParams *[]interface{}, root *xml.Node) {
	// 获取目标需要循环的对象，默认为list
	collectionKey, ok := node.Attrs["collection"]
	if !ok {
		collectionKey = "list"
	}
	collection, ok := params[collectionKey]
	if !ok {
		panic("集合不存在")
		// return errors.New("集合不存在")
	}
	// 必须是slice或者array
	collectionType := reflect.TypeOf(collection)
	if collectionType.Kind() != reflect.Slice && collectionType.Kind() != reflect.Array {
		panic("集合类型错误")
		// return errors.New("集合类型错误")
	}

	// 获取映射的key，默认为item
	mapKey, ok := node.Attrs["item"]
	if !ok {
		mapKey = "item"
	}
	// 获取分隔符，默认为逗号
	separator, ok := node.Attrs["separator"]
	if !ok {
		separator = ","
	}
	// 获取open和close，默认为空即可
	open, ok := node.Attrs["open"]
	if !ok {
		open = ""
	}
	close, ok := node.Attrs["close"]
	if !ok {
		close = ""
	}
	if open != "" {
		builder.WriteString(open)
	}
	// 拼接字符串
	// 将 collection 转换为数组
	var childBuilder0 strings.Builder
	var childBuilder1 strings.Builder
	collectionValue := reflect.ValueOf(collection)
	collectionLength := collectionValue.Len()
	for i := 0; i < collectionLength; i++ {
		childBuilder0.Reset()
		params[mapKey] = collectionValue.Index(i).Interface()
		// 获取map的key
		for _, child := range node.Children {
			HandleNode(&childBuilder0, child, params, resultParams, root)
		}
		if childBuilder0.Len() > 0 {
			childBuilder0.WriteString(separator)
		}
		childBuilder1.WriteString(childBuilder0.String())
	}
	// 去除childbuilder1最后的separator
	childStr := childBuilder1.String()
	if len(childStr) > 0 {
		// 去除childStr最后的separator
		childStr = strings.TrimSuffix(childStr, separator)
		builder.WriteString(childStr)
	}
	if close != "" {
		builder.WriteString(close)
	}
}

func handleWhereStatement(builder *strings.Builder, node *xml.Node, params map[string]interface{}, resultParams *[]interface{}, root *xml.Node) {

	// 移除第一个 "AND" 或 "OR"
	sqlBuilder := strings.Builder{}
	isFirstCondition := true
	for _, child := range node.Children {
		childBuilder := &strings.Builder{}
		HandleNode(childBuilder, child, params, resultParams, root)

		childSQL := strings.TrimSpace(childBuilder.String())

		if strings.HasPrefix(strings.ToUpper(childSQL), "AND") || strings.HasPrefix(strings.ToUpper(childSQL), "OR") {
			if isFirstCondition {
				childSQL = strings.TrimSpace(childSQL[3:])
				isFirstCondition = false
			}
		}

		sqlBuilder.WriteString(childSQL)
		sqlBuilder.WriteString(" ")
	}
	childSql := strings.TrimSpace(sqlBuilder.String())
	if childSql != "" {
		builder.WriteString(" where ")
		builder.WriteString(childSql)
		builder.WriteByte(' ')
	}
}

func handleSetStatement(builder *strings.Builder, node *xml.Node, params map[string]interface{}, resultParams *[]interface{}, root *xml.Node) {
	builder.WriteString(" set ")
	var childBuilder strings.Builder
	// 移除末尾的逗号
	for _, child := range node.Children {
		HandleNode(&childBuilder, child, params, resultParams, root)
	}
	childSQL := strings.TrimSpace(childBuilder.String())
	if strings.HasSuffix(childSQL, ",") {
		childSQL = childSQL[:len(childSQL)-1]
	}
	builder.WriteString(childSQL)
	builder.WriteString(" ")
}

func handleSqlStatement(builder *strings.Builder, node *xml.Node, params map[string]interface{}, resultParams *[]interface{}, root *xml.Node) {
	for _, child := range node.Children {
		HandleNode(builder, child, params, resultParams, root)
	}
}

func handleIncludeStatement(builder *strings.Builder, node *xml.Node, params map[string]interface{}, resultParams *[]interface{}, root *xml.Node) {
	// 获取include的文件名
	refid, ok := node.Attrs["refid"]
	if !ok {
		return
	}
	// 查找root中是否有符合id的节点
	for _, child := range root.Children {
		if child.Attrs["id"] == refid {
			// 不能是自身
			if child == node {
				continue
			}
			// 获取include的文件内容
			HandleNode(builder, child, params, resultParams, root)
			break
		}
	}
}

func handleCustomStatement(builder *strings.Builder, node *xml.Node, params map[string]interface{}, resultParams *[]interface{}, root *xml.Node) {
	handler, ok := plugin.GetTagHandler(node.Name)
	if !ok {
		panic(fmt.Sprintf("未找到标签处理器: %s", node.Name))
	}
	handler(builder, node, params, resultParams, root)
}

// 处理文本节点
func handleText(builder *strings.Builder, node *xml.Node, params map[string]interface{}, resultParams *[]interface{}) {
	text := node.Text
	if text == "" {
		return
	}
	// 使用正则表达式匹配 #{abc.xxx} 和 ${abc.xxx} 格式的内容
	re := regexp.MustCompile(`(#|\$)\{([^}]+)\}`)

	text = re.ReplaceAllStringFunc(text, func(match string) string {
		if strings.HasPrefix(match, "${") {
			// 处理 ${} 格式，直接拼接
			key := strings.Trim(match, "${}")
			value := runner.GetValue(key, params)
			return fmt.Sprintf("%v", value)
		} else {
			// 处理 #{} 格式，使用参数化查询
			// 特殊情况，如果key为$AUTO，则自动生成id
			value := getValueByBlock(match, params)
			if value == "$AUTO" {
				return "DEFAULT"
			}
			*resultParams = append(*resultParams, value)
			return "?"
		}
	})

	builder.WriteString(text)
}

// 根据key获取值
// 处理 #{abc.xxx} 格式的内容
// todo: 处理 ${abc.xxx} 格式的内容 即不处理内容直接输出
func getValueByBlock(key string, params map[string]interface{}) interface{} {
	key = strings.Trim(key, "#{}")
	// 处理三元表达式的情况
	if strings.Contains(key, "?") || strings.Contains(key, "(") {
		value := runner.EvaluateExpression(key, params)
		return value
		//return fmt.Sprintf("%v", value)
	} else {
		value := runner.GetValue(key, params)
		return value
		//return fmt.Sprintf("%v", value)
	}
}

func getValueByDollarBlock(key string, params map[string]interface{}) string {
	key = strings.Trim(key, "${}")
	// 处理三元表达式
	if strings.Contains(key, "?") {
		value := runner.EvaluateExpression(key, params)
		return fmt.Sprintf("%v", value)
	} else {
		value := runner.GetValue(key, params)
		return fmt.Sprintf("%v", value)
	}
}

func extractObject(v interface{}, params map[string]interface{}) {
	typeOfV := reflect.TypeOf(v)
	if typeOfV.Kind() == reflect.Map {
		for mk, mv := range v.(map[string]interface{}) {
			params[mk] = mv
		}
	} else if typeOfV.Kind() == reflect.Ptr && typeOfV.Elem().Kind() == reflect.Struct {
		extractStructFields(reflect.ValueOf(v).Elem(), params)
	} else if typeOfV.Kind() == reflect.Struct {
		extractStructFields(reflect.ValueOf(v), params)
	}
}

func extractStructFields(structValue reflect.Value, params map[string]interface{}) {
	// 如果是结构体指针，展开结构体字段
	structType := structValue.Type()
	for i := 0; i < structValue.NumField(); i++ {
		field := structType.Field(i)
		fieldValue := structValue.Field(i)

		// 优先使用vo tag
		key := field.Tag.Get("vo")
		if key == "" {
			key = field.Tag.Get("json")
		}
		if key == "" {
			key = field.Name
		}

		params[key] = fieldValue.Interface()
	}
}
