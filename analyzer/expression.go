package analyzer

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

// TokenType 定义
const (
	Identifier  TokenType = "Identifier"
	Number      TokenType = "Number"
	Operator    TokenType = "Operator"
	Parenthesis TokenType = "Parenthesis"
)

type TokenType string
type Token struct {
	Type  TokenType
	Value string
}
type ASTNode struct {
	Type  string
	Value string
	Left  *ASTNode
	Right *ASTNode
}
type AST struct {
	Root *ASTNode
}

func EvaluateExpression(expr string, params map[string]interface{}) bool {
	// 词法分析
	tokens := lexicalAnalysis(expr)

	// 语法分析和规约
	ast := syntaxAnalysis(tokens)

	printAST(ast)

	// 计算表达式
	val := evaluateAST(ast, params)
	valBool, ok := val.(bool)
	if ok {
		return valBool
	}
	return false
}

func printAST(ast *ASTNode) {
	if ast == nil {
		return
	}
	fmt.Printf("%s: %s\n", ast.Type, ast.Value)
	printAST(ast.Left)
	printAST(ast.Right)
}

// 词法分析器
func lexicalAnalysis(expr string) []Token {
	tokens := make([]Token, 0)
	var currentToken string
	var tokenType TokenType

	for i := 0; i < len(expr); i++ {
		char := expr[i]

		switch {
		case isWhitespace(char):
			if currentToken != "" {
				tokens = append(tokens, Token{Type: tokenType, Value: currentToken})
				currentToken = ""
			}
		case isLetter(char) || char == '.':
			if currentToken == "" {
				tokenType = Identifier
			}
			currentToken += string(char)
		case isDigit(char):
			if currentToken == "" {
				tokenType = Number
			}
			currentToken += string(char)
		case isOperator(char):
			if currentToken != "" {
				tokens = append(tokens, Token{Type: tokenType, Value: currentToken})
				currentToken = ""
			}
			if i+1 < len(expr) && isCompoundOperator(char, expr[i+1]) {
				tokens = append(tokens, Token{Type: Operator, Value: string(char) + string(expr[i+1])})
				i++
			} else {
				tokens = append(tokens, Token{Type: Operator, Value: string(char)})
			}
		case char == '(' || char == ')':
			if currentToken != "" {
				tokens = append(tokens, Token{Type: tokenType, Value: currentToken})
				currentToken = ""
			}
			tokens = append(tokens, Token{Type: Parenthesis, Value: string(char)})
		}
	}

	if currentToken != "" {
		tokens = append(tokens, Token{Type: tokenType, Value: currentToken})
	}

	return tokens
}

func isCompoundOperator(first, second byte) bool {
	return (first == '=' && second == '=') ||
		(first == '!' && second == '=') ||
		(first == '>' && second == '=') ||
		(first == '<' && second == '=')
}

// 语法分析器
func syntaxAnalysis(tokens []Token) *ASTNode {
	var pos int
	return parseExpression(tokens, &pos)
}

func parseExpression(tokens []Token, pos *int) *ASTNode {
	node := parseTerm(tokens, pos)
	for *pos < len(tokens) && (tokens[*pos].Value == "||") {
		op := tokens[*pos]
		*pos++
		right := parseTerm(tokens, pos)
		node = &ASTNode{
			Type:  "BinaryOp",
			Value: op.Value,
			Left:  node,
			Right: right,
		}
	}
	return node
}

func parseTerm(tokens []Token, pos *int) *ASTNode {
	node := parseFactor(tokens, pos)
	for *pos < len(tokens) && (tokens[*pos].Value == "&&") {
		op := tokens[*pos]
		*pos++
		right := parseFactor(tokens, pos)
		node = &ASTNode{
			Type:  "BinaryOp",
			Value: op.Value,
			Left:  node,
			Right: right,
		}
	}
	return node
}

func parseFactor(tokens []Token, pos *int) *ASTNode {
	node := parsePrimary(tokens, pos)
	for *pos < len(tokens) && isComparisonOperator(tokens[*pos].Value) {
		op := tokens[*pos]
		*pos++
		right := parsePrimary(tokens, pos)
		node = &ASTNode{
			Type:  "BinaryOp",
			Value: op.Value,
			Left:  node,
			Right: right,
		}
	}
	return node
}

func parsePrimary(tokens []Token, pos *int) *ASTNode {
	token := tokens[*pos]
	*pos++
	switch token.Type {
	case Operator:
		if token.Value == "!" {
			operand := parsePrimary(tokens, pos)
			return &ASTNode{
				Type:  "UnaryOp",
				Value: token.Value,
				Left:  operand,
			}
		}
	case Identifier:
		return &ASTNode{Type: "Identifier", Value: token.Value}
	case Number:
		return &ASTNode{Type: "Literal", Value: token.Value}
	case Parenthesis:
		if token.Value == "(" {
			node := parseExpression(tokens, pos)
			if tokens[*pos].Value != ")" {
				panic("缺少右括号")
			}
			*pos++
			return node
		}
	}
	panic(fmt.Sprintf("无法解析的Token: %v", token))
}

func isComparisonOperator(op string) bool {
	return op == "==" || op == "!=" || op == ">" || op == "<" || op == ">=" || op == "<="
}

// 辅助函数：获取操作符优先级
func precedence(op string) int {
	switch op {
	case "&&", "||":
		return 1
	case "==", "!=", ">", "<", ">=", "<=":
		return 2
	default:
		return 0
	}
}

// 辅助函数：将interface{}转换为ASTNode
func nodeFromInterface(i interface{}) *ASTNode {
	switch v := i.(type) {
	case *ASTNode:
		return v
	case Token:
		return &ASTNode{
			Type:  string(v.Type),
			Value: v.Value,
		}
	default:
		return nil
	}
}

func evaluateAST(ast *ASTNode, params map[string]interface{}) interface{} {
	// 遍历AST并计算表达式结果
	// 返回布尔值结果
	switch ast.Type {
	case "BinaryOp":
		left := evaluateAST(ast.Left, params)
		right := evaluateAST(ast.Right, params)
		switch ast.Value {
		case "&&":
			return left.(bool) && right.(bool)
		case "||":
			return left.(bool) || right.(bool)
		case "==":
			return left == right
		case "!=":
			return left != right
		case ">":
			return compareValues(left, right, ">")
		case "<":
			return compareValues(left, right, "<")
		case ">=":
			return compareValues(left, right, ">=")
		case "<=":
			return compareValues(left, right, "<=")
		default:
			panic(fmt.Sprintf("不支持的操作符: %s", ast.Value))
		}
	case "UnaryOp":
		operand := evaluateAST(ast.Left, params)
		switch ast.Value {
		case "!":
			return !operand.(bool)
		default:
			panic(fmt.Sprintf("不支持的操作符: %s", ast.Value))
		}
	case "Identifier":
		value := getValue(ast.Value, params)
		return value
	case "Literal":
		return ast.Value
	default:
		panic(fmt.Sprintf("未知的节点类型: %s", ast.Type))
	}
}

// 辅助函数：判断字符是否为空白字符
func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// 辅助函数：判断字符是否为字母
func isLetter(c byte) bool {
	return unicode.IsLetter(rune(c))
}

// 辅助函数：判断字符是否为数字
func isDigit(c byte) bool {
	return unicode.IsDigit(rune(c))
}

// 辅助函数：判断字符是否为操作符
func isOperator(c byte) bool {
	return c == '&' || c == '|' || c == '=' || c == '!' || c == '>' || c == '<'
}

// 辅助函数：比较两个值
func compareValues(left, right interface{}, op string) bool {
	// 都是布尔的情况
	if left == true && right == true {
		return true
	}
	if left == false && right == false {
		return true
	}
	if left == false || right == false {
		return false
	}
	// 实现比较逻辑
	leftFloat, leftOk := toFloat64(left)
	rightFloat, rightOk := toFloat64(right)

	if !leftOk || !rightOk {
		panic(fmt.Sprintf("无法比较值: %v 和 %v", left, right))
	}

	switch op {
	case ">":
		return leftFloat > rightFloat
	case "<":
		return leftFloat < rightFloat
	case ">=":
		return leftFloat >= rightFloat
	case "<=":
		return leftFloat <= rightFloat
	default:
		panic(fmt.Sprintf("不支持的比较操作符: %s", op))
	}
}

// 辅助函数：将接口类型转换为float64
func toFloat64(v interface{}) (float64, bool) {
	switch value := v.(type) {
	case float64:
		return value, true
	case float32:
		return float64(value), true
	case int:
		return float64(value), true
	case int64:
		return float64(value), true
	case string:
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// 辅助函数：从参数映射中获取值
func getValueFromMap(key string, params map[string]interface{}) interface{} {
	if value, ok := params[key]; ok {
		return value
	}
	return nil
}

func getValue(key string, params interface{}) interface{} {
	keys := strings.Split(key, ".")
	value := params
	for _, k := range keys {
		switch v := value.(type) {
		case map[string]interface{}:
			if val, ok := v[k]; ok {
				value = val
			} else {
				return nil
				// panic(fmt.Sprintf("属性 %s 不存在", k))
				// return "", fmt.Errorf("属性 %s 不存在", k)
			}
		default:
			rv := reflect.ValueOf(v)
			if rv.Kind() == reflect.Struct {
				// 优先尝试读取 vo 标签
				field, found := rv.Type().FieldByNameFunc(func(fieldName string) bool {
					field, _ := rv.Type().FieldByName(fieldName)
					tag := field.Tag.Get("vo")
					return tag == k || fieldName == k
				})

				if found {
					value = rv.FieldByIndex(field.Index).Interface()
				} else {
					return nil
				}
			} else {
				return nil
			}
		}
	}
	return value
}
