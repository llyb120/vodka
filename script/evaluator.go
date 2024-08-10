package script

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

// TokenType 定义
type TokenType int

const (
	NUMBER TokenType = iota
	PLUS
	MINUS
	MULTIPLY
	DIVIDE
	ASSIGN
	IDENTIFIER
	LPAREN
	RPAREN
	COMMA
	FUNCTION
	RETURN
	QUESTION
	COLON
	EOF
)

// Token 结构
type Token struct {
	Type  TokenType
	Value string
}

// AST 节点接口
type Node interface {
	Evaluate(env *Environment) interface{}
}

// 数字节点
type NumberNode struct {
	Value float64
}

func (n *NumberNode) Evaluate(env *Environment) interface{} {
	return n.Value
}

// 二元操作节点
type BinaryOpNode struct {
	Left     Node
	Operator TokenType
	Right    Node
}

func (n *BinaryOpNode) Evaluate(env *Environment) interface{} {
	left := n.Left.Evaluate(env).(float64)
	right := n.Right.Evaluate(env).(float64)
	switch n.Operator {
	case PLUS:
		return left + right
	case MINUS:
		return left - right
	case MULTIPLY:
		return left * right
	case DIVIDE:
		return left / right
	}
	return 0
}

// 变量节点
type VariableNode struct {
	Name string
}

func (n *VariableNode) Evaluate(env *Environment) interface{} {
	return env.Get(n.Name)
}

// 赋值节点
type AssignNode struct {
	Name  string
	Value Node
}

func (n *AssignNode) Evaluate(env *Environment) interface{} {
	value := n.Value.Evaluate(env)
	env.Set(n.Name, value)
	return value
}

// 函数定义节点
type FunctionDefNode struct {
	Name       string
	Parameters []string
	Body       Node
}

func (n *FunctionDefNode) Evaluate(env *Environment) interface{} {
	env.SetFunction(n.Name, n)
	return nil
}

// 函数调用节点
type FunctionCallNode struct {
	Name      string
	Arguments []Node
}

func (n *FunctionCallNode) Evaluate(env *Environment) interface{} {
	function := env.GetFunction(n.Name)
	localEnv := NewEnvironment(env)
	for i, param := range function.Parameters {
		localEnv.Set(param, n.Arguments[i].Evaluate(env))
	}
	return function.Body.Evaluate(localEnv)
}

// 三元表达式节点
type TernaryNode struct {
	Condition Node
	TrueExpr  Node
	FalseExpr Node
}

func (n *TernaryNode) Evaluate(env *Environment) interface{} {
	condition := n.Condition.Evaluate(env).(float64)
	if condition != 0 {
		return n.TrueExpr.Evaluate(env)
	}
	return n.FalseExpr.Evaluate(env)
}

// 环境
type Environment struct {
	variables map[string]interface{}
	functions map[string]*FunctionDefNode
	parent    *Environment
}

func NewEnvironment(parent *Environment) *Environment {
	return &Environment{
		variables: make(map[string]interface{}),
		functions: make(map[string]*FunctionDefNode),
		parent:    parent,
	}
}

func (e *Environment) Get(name string) interface{} {
	if value, ok := e.variables[name]; ok {
		return value
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return 0
}

func (e *Environment) Set(name string, value interface{}) {
	e.variables[name] = value
}

func (e *Environment) SetFunction(name string, fn *FunctionDefNode) {
	e.functions[name] = fn
}

func (e *Environment) GetFunction(name string) *FunctionDefNode {
	if fn, ok := e.functions[name]; ok {
		return fn
	}
	if e.parent != nil {
		return e.parent.GetFunction(name)
	}
	return nil
}

// 新增语句节点类型
type StatementNode interface {
	Node
	statementNode()
}

// 表达式语句节点
type ExpressionStatementNode struct {
	Expression Node
}

func (n *ExpressionStatementNode) Evaluate(env *Environment) interface{} {
	return n.Expression.Evaluate(env)
}

func (n *ExpressionStatementNode) statementNode() {}

// 程序节点
type ProgramNode struct {
	Statements []StatementNode
}

func (n *ProgramNode) Evaluate(env *Environment) interface{} {
	var result interface{}
	for _, stmt := range n.Statements {
		result = stmt.Evaluate(env)
	}
	return result
}

// 更新词法分析器以处理多行输入
func tokenize(input string) []Token {
	tokens := []Token{}
	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		word := scanner.Text()
		switch word {
		case "+":
			tokens = append(tokens, Token{PLUS, word})
		case "-":
			tokens = append(tokens, Token{MINUS, word})
		case "*":
			tokens = append(tokens, Token{MULTIPLY, word})
		case "/":
			tokens = append(tokens, Token{DIVIDE, word})
		case "=":
			tokens = append(tokens, Token{ASSIGN, word})
		case "(":
			tokens = append(tokens, Token{LPAREN, word})
		case ")":
			tokens = append(tokens, Token{RPAREN, word})
		case ",":
			tokens = append(tokens, Token{COMMA, word})
		case "function":
			tokens = append(tokens, Token{FUNCTION, word})
		case "return":
			tokens = append(tokens, Token{RETURN, word})
		case "?":
			tokens = append(tokens, Token{QUESTION, word})
		case ":":
			tokens = append(tokens, Token{COLON, word})
		default:
			if _, err := strconv.ParseFloat(word, 64); err == nil {
				tokens = append(tokens, Token{NUMBER, word})
			} else {
				tokens = append(tokens, Token{IDENTIFIER, word})
			}
		}
	}
	tokens = append(tokens, Token{EOF, ""})
	return tokens
}

// 更新语法分析器
type Parser struct {
	tokens  []Token
	current int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens, current: 0}
}

func (p *Parser) parse() *ProgramNode {
	statements := []StatementNode{}
	for !p.isAtEnd() {
		statements = append(statements, p.statement())
	}
	return &ProgramNode{Statements: statements}
}

func (p *Parser) statement() StatementNode {
	return &ExpressionStatementNode{Expression: p.expression()}
}

func (p *Parser) expression() Node {
	return p.assignment()
}

func (p *Parser) assignment() Node {
	if p.match(IDENTIFIER) && p.peek().Type == ASSIGN {
		name := p.previous().Value
		p.consume(ASSIGN)
		value := p.expression()
		return &AssignNode{Name: name, Value: value}
	}
	return p.ternary()
}

func (p *Parser) ternary() Node {
	expr := p.additive()
	if p.match(QUESTION) {
		trueExpr := p.expression()
		p.consume(COLON)
		falseExpr := p.expression()
		return &TernaryNode{Condition: expr, TrueExpr: trueExpr, FalseExpr: falseExpr}
	}
	return expr
}

func (p *Parser) additive() Node {
	expr := p.multiplicative()
	for p.match(PLUS, MINUS) {
		operator := p.previous().Type
		right := p.multiplicative()
		expr = &BinaryOpNode{Left: expr, Operator: operator, Right: right}
	}
	return expr
}

func (p *Parser) multiplicative() Node {
	expr := p.primary()
	for p.match(MULTIPLY, DIVIDE) {
		operator := p.previous().Type
		right := p.primary()
		expr = &BinaryOpNode{Left: expr, Operator: operator, Right: right}
	}
	return expr
}

func (p *Parser) primary() Node {
	if p.match(NUMBER) {
		value, _ := strconv.ParseFloat(p.previous().Value, 64)
		return &NumberNode{Value: value}
	}
	if p.match(IDENTIFIER) {
		name := p.previous().Value
		if p.match(LPAREN) {
			args := []Node{}
			if !p.check(RPAREN) {
				for {
					args = append(args, p.expression())
					if !p.match(COMMA) {
						break
					}
				}
			}
			p.consume(RPAREN)
			return &FunctionCallNode{Name: name, Arguments: args}
		}
		return &VariableNode{Name: name}
	}
	if p.match(LPAREN) {
		expr := p.expression()
		p.consume(RPAREN)
		return expr
	}
	if p.match(FUNCTION) {
		return p.functionDefinition()
	}
	panic("Unexpected token")
}

func (p *Parser) functionDefinition() Node {
	name := p.consume(IDENTIFIER).Value
	p.consume(LPAREN)
	params := []string{}
	if !p.check(RPAREN) {
		for {
			params = append(params, p.consume(IDENTIFIER).Value)
			if !p.match(COMMA) {
				break
			}
		}
	}
	p.consume(RPAREN)
	body := p.expression()
	return &FunctionDefNode{Name: name, Parameters: params, Body: body}
}

func (p *Parser) match(types ...TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) check(t TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == t
}

func (p *Parser) advance() Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

func (p *Parser) isAtEnd() bool {
	return p.peek().Type == EOF
}

func (p *Parser) peek() Token {
	return p.tokens[p.current]
}

func (p *Parser) previous() Token {
	return p.tokens[p.current-1]
}

func (p *Parser) consume(t TokenType) Token {
	if p.check(t) {
		return p.advance()
	}
	panic(fmt.Sprintf("Expected %v", t))
}

func main() {
	env := NewEnvironment(nil)

	inputs := []string{
		"x = 5 + 3 * 2",
		"function add(a, b) a + b",
		"add(x, 10)",
		"y = 1 ? 20 : 30",
	}

	for _, input := range inputs {
		tokens := tokenize(input)
		parser := NewParser(tokens)
		ast := parser.parse()
		result := ast.Evaluate(env)
		fmt.Printf("输入: %s\n结果: %v\n\n", input, result)
	}
}

func Evaluate(input string) (interface{}, error) {
	env := NewEnvironment(nil)
	tokens := tokenize(input)
	parser := NewParser(tokens)
	ast := parser.parse()
	return ast.Evaluate(env), nil
}
