package xml

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

type TokenType int

const (
	Tag = iota
	Text
	//TokenError TokenType = iota
	//TokenEOF
	//TokenText
	//TokenOpenTag
	//TokenCloseTag
	//TokenSelfCloseTag
	//TokenAttrName
	//TokenAttrValue
)

type Token struct {
	Type  TokenType
	Value string
}

type Parser struct {
	input  string
	start  int
	pos    int
	width  int
	tokens chan Token
}

func NewParser(input string) *Parser {
	parser := &Parser{
		input:  input,
		tokens: make(chan Token),
		start:  0,
		pos:    0,
		width:  0,
	}
	return parser
}

func (p *Parser) Parse() (*Node, error) {
	stack := make([]*Node, 0)
	var root *Node
	var builder strings.Builder

	handleTextNode := func() {
		if builder.Len() == 0 {
			return
		}
		// trim后的空节点不要的干活
		text := builder.String()
		//fmt.Println("text:", text)
		// 处理text的缓冲
		if strings.TrimSpace(text) != "" && len(stack) > 0 {
			pNode := stack[len(stack)-1]
			node := &Node{Type: Text, Text: text}
			pNode.Children = append(pNode.Children, node)
			builder.Reset()
		}
	}

LOOP:
	for {
		switch r := p.next(); {
		case r == 0:
			handleTextNode()
			break LOOP
		case r == '<':
			// 如果不在标签中
			if p.isLetter(p.peek()) || p.peek() == '!' || p.peek() == '?' || p.peek() == '/' {
				handleTextNode()
				if p.peek() == '/' {
					// 如果是结束标签
					p.next()
					token := p.readUntil('>')
					// token必须和栈顶元素的name相同
					if len(stack) > 0 {
						pNode := stack[len(stack)-1]
						if pNode.Name != strings.ToUpper(token) {
							return nil, errors.New("标签不匹配")
						}
						// 弹出栈顶元素
						stack = stack[:len(stack)-1]
					}
					// 跳过
					p.next()
				} else if p.peek() == '?' {
					p.readUntil('>')
					p.next()
				} else if p.peek() == '!' {
					p.readUntil('>')
					p.next()
				} else {
					// 如果不是空格，则开始进入标签
					token := p.readUntil(' ', '\t', '>')
					node := &Node{Type: Tag, Name: strings.ToUpper(token), Attrs: make(map[string]string), Children: make([]*Node, 0)}
					p.readAttributes(node)
					// 跳过结束标签
					p.next()
					// 放入栈中
					if len(stack) > 0 {
						pNode := stack[len(stack)-1]
						pNode.Children = append(pNode.Children, node)
					} else {
						root = node
					}
					stack = append(stack, node)
				}
			} else {
				builder.WriteRune(r)
			}

		default:
			builder.WriteRune(r)
		}
	}
	return root, nil
}

func (p *Parser) readAttributes(node *Node) {
	// 读取所有可能出现的属性
	attrKey := ""
	attrValue := ""
	for {
		p.consumeWhitespace()
		if p.peek() == '>' || p.peek() == 0 {
			break
		}
		attrKey = p.readUntil('=', ' ', '\t', '>')
		if p.peek() == '=' {
			p.next() // 跳过等号
			p.consumeWhitespace()
			if p.peek() == '"' || p.peek() == '\'' {
				quote := p.next() // 跳过开始引号
				attrValue = p.readUntil(quote)
				p.next() // 跳过结束引号
			} else {
				attrValue = p.readUntil(' ', '\t', '>')
			}
		} else {
			attrValue = "" // 独立单词属性设置空值
		}
		node.Attrs[attrKey] = attrValue
	}
}

func (p *Parser) next() rune {
	if p.pos >= len(p.input) {
		return 0
	}
	r := rune(p.input[p.pos])
	p.pos += 1
	return r
}

func (p *Parser) peek() rune {
	if p.pos >= len(p.input) {
		return 0
	}
	return rune(p.input[p.pos])
}

func (p *Parser) consumeWhitespace() {
	for {
		if p.isSpace(p.peek()) {
			p.next()
		} else {
			break
		}
	}
}

func (p *Parser) backup() {
	p.pos -= 1
}

func (p *Parser) isSpace(r rune) bool {
	return unicode.IsSpace(r)
}

func (p *Parser) isLetter(r rune) bool {
	return unicode.IsLetter(r)
}

func (p *Parser) readUntil(end ...rune) string {
	start := p.pos
	for {
		r := p.next()
		for _, e := range end {
			if r == e {
				p.backup()
				return p.input[start:p.pos]
			}
		}
		if r == 0 {
			break
		}
	}
	return p.input[start:p.pos]
}

type Node struct {
	Type     TokenType
	Name     string
	Attrs    map[string]string
	Children []*Node
	Text     string
}

func main() {
	xml := `<note id test=1 a='1' b="2" c="b > 2" d="e<2">
	<to>Tove</to>
	<from>Jani</from>
	<heading>Reminder</heading>
	<body>Don't forget me this weekend!</body>
	</note>`

	// lexer := NewLexer(xml)
	// for {
	// 	token := lexer.NextToken()
	// 	fmt.Println(token)
	// 	if token.Type == TokenEOF {
	// 		break
	// 	}
	// }
	// root := ParseXML(xml)
	// fmt.Println(root)
	parser := NewParser(xml)
	root, _ := parser.Parse()
	fmt.Print(root)
}
