package spec

import (
	"fmt"
)

type cmdExpr interface {
	isCmdExpr()
}

type cmdStringExpr struct {
	Value string
}

func (*cmdStringExpr) isCmdExpr() {}

type cmdNumberExpr struct {
	Value string
}

func (*cmdNumberExpr) isCmdExpr() {}

type cmdBoolExpr struct {
	Value bool
}

func (*cmdBoolExpr) isCmdExpr() {}

type cmdNullExpr struct{}

func (*cmdNullExpr) isCmdExpr() {}

type cmdIdentExpr struct {
	Name string
}

func (*cmdIdentExpr) isCmdExpr() {}

type cmdMemberExpr struct {
	Target cmdExpr
	Name   string
}

func (*cmdMemberExpr) isCmdExpr() {}

type cmdCallExpr struct {
	Callee cmdExpr
	Args   []cmdExpr
}

func (*cmdCallExpr) isCmdExpr() {}

type cmdTypeofExpr struct {
	Expr cmdExpr
}

func (*cmdTypeofExpr) isCmdExpr() {}

type cmdBinaryExpr struct {
	Op    string
	Left  cmdExpr
	Right cmdExpr
}

func (*cmdBinaryExpr) isCmdExpr() {}

type cmdConditionalExpr struct {
	Cond      cmdExpr
	WhenTrue  cmdExpr
	WhenFalse cmdExpr
}

func (*cmdConditionalExpr) isCmdExpr() {}

type cmdTokenKind int

const (
	cmdTokenEOF cmdTokenKind = iota
	cmdTokenIdent
	cmdTokenString
	cmdTokenNumber
	cmdTokenPlus
	cmdTokenQuestion
	cmdTokenColon
	cmdTokenLParen
	cmdTokenRParen
	cmdTokenDot
	cmdTokenComma
	cmdTokenEqEq
	cmdTokenSemicolon
)

type cmdToken struct {
	Kind cmdTokenKind
	Lit  string
}

type cmdLexer struct {
	src string
	pos int
}

func newCmdLexer(src string) *cmdLexer {
	return &cmdLexer{src: src}
}

func (l *cmdLexer) nextToken() (cmdToken, error) {
	l.skipSpaces()
	if l.pos >= len(l.src) {
		return cmdToken{Kind: cmdTokenEOF}, nil
	}

	ch := l.src[l.pos]
	switch ch {
	case '\'':
		return l.readString()
	case '+':
		l.pos++
		return cmdToken{Kind: cmdTokenPlus, Lit: "+"}, nil
	case '?':
		l.pos++
		return cmdToken{Kind: cmdTokenQuestion, Lit: "?"}, nil
	case ':':
		l.pos++
		return cmdToken{Kind: cmdTokenColon, Lit: ":"}, nil
	case '(':
		l.pos++
		return cmdToken{Kind: cmdTokenLParen, Lit: "("}, nil
	case ')':
		l.pos++
		return cmdToken{Kind: cmdTokenRParen, Lit: ")"}, nil
	case '.':
		l.pos++
		return cmdToken{Kind: cmdTokenDot, Lit: "."}, nil
	case ',':
		l.pos++
		return cmdToken{Kind: cmdTokenComma, Lit: ","}, nil
	case ';':
		l.pos++
		return cmdToken{Kind: cmdTokenSemicolon, Lit: ";"}, nil
	case '=':
		if l.pos+1 < len(l.src) && l.src[l.pos+1] == '=' {
			l.pos += 2
			return cmdToken{Kind: cmdTokenEqEq, Lit: "=="}, nil
		}
		return cmdToken{}, fmt.Errorf("unexpected '=' at %d", l.pos)
	}

	if isDigit(ch) {
		return l.readNumber(), nil
	}
	if isIdentStart(ch) {
		return l.readIdent(), nil
	}

	return cmdToken{}, fmt.Errorf("unexpected character %q at %d", ch, l.pos)
}

func (l *cmdLexer) skipSpaces() {
	for l.pos < len(l.src) {
		switch l.src[l.pos] {
		case ' ', '\t', '\n', '\r':
			l.pos++
		default:
			return
		}
	}
}

func (l *cmdLexer) readString() (cmdToken, error) {
	l.pos++
	out := make([]byte, 0, 16)
	for l.pos < len(l.src) {
		ch := l.src[l.pos]
		l.pos++
		if ch == '\'' {
			return cmdToken{Kind: cmdTokenString, Lit: string(out)}, nil
		}
		if ch == '\\' {
			if l.pos >= len(l.src) {
				return cmdToken{}, fmt.Errorf("unterminated escape sequence")
			}
			esc := l.src[l.pos]
			l.pos++
			switch esc {
			case '\\', '\'', '"':
				out = append(out, esc)
			case 'n':
				out = append(out, '\n')
			case 'r':
				out = append(out, '\r')
			case 't':
				out = append(out, '\t')
			default:
				out = append(out, esc)
			}
			continue
		}
		out = append(out, ch)
	}
	return cmdToken{}, fmt.Errorf("unterminated string literal")
}

func (l *cmdLexer) readNumber() cmdToken {
	start := l.pos
	for l.pos < len(l.src) && isDigit(l.src[l.pos]) {
		l.pos++
	}
	if l.pos < len(l.src) && l.src[l.pos] == '.' {
		l.pos++
		for l.pos < len(l.src) && isDigit(l.src[l.pos]) {
			l.pos++
		}
	}
	return cmdToken{Kind: cmdTokenNumber, Lit: l.src[start:l.pos]}
}

func (l *cmdLexer) readIdent() cmdToken {
	start := l.pos
	l.pos++
	for l.pos < len(l.src) && isIdentPart(l.src[l.pos]) {
		l.pos++
	}
	return cmdToken{Kind: cmdTokenIdent, Lit: l.src[start:l.pos]}
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isIdentStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' || ch == '$'
}

func isIdentPart(ch byte) bool {
	return isIdentStart(ch) || isDigit(ch)
}

type cmdExprParser struct {
	lexer *cmdLexer
	cur   cmdToken
}

func parseCmdExpression(expr string) (cmdExpr, error) {
	p := &cmdExprParser{lexer: newCmdLexer(expr)}
	if err := p.advance(); err != nil {
		return nil, err
	}
	root, err := p.parseConditional()
	if err != nil {
		return nil, err
	}
	if p.cur.Kind == cmdTokenSemicolon {
		if err := p.advance(); err != nil {
			return nil, err
		}
	}
	if p.cur.Kind != cmdTokenEOF {
		return nil, fmt.Errorf("unexpected token %q", p.cur.Lit)
	}
	return root, nil
}

func (p *cmdExprParser) advance() error {
	tok, err := p.lexer.nextToken()
	if err != nil {
		return err
	}
	p.cur = tok
	return nil
}

func (p *cmdExprParser) parseConditional() (cmdExpr, error) {
	cond, err := p.parseEquality()
	if err != nil {
		return nil, err
	}
	if p.cur.Kind != cmdTokenQuestion {
		return cond, nil
	}
	if err := p.advance(); err != nil {
		return nil, err
	}
	whenTrue, err := p.parseConditional()
	if err != nil {
		return nil, err
	}
	if p.cur.Kind != cmdTokenColon {
		return nil, fmt.Errorf("expected ':' in conditional expression")
	}
	if err := p.advance(); err != nil {
		return nil, err
	}
	whenFalse, err := p.parseConditional()
	if err != nil {
		return nil, err
	}
	return &cmdConditionalExpr{
		Cond:      cond,
		WhenTrue:  whenTrue,
		WhenFalse: whenFalse,
	}, nil
}

func (p *cmdExprParser) parseEquality() (cmdExpr, error) {
	left, err := p.parseAdditive()
	if err != nil {
		return nil, err
	}
	for p.cur.Kind == cmdTokenEqEq {
		if err := p.advance(); err != nil {
			return nil, err
		}
		right, err := p.parseAdditive()
		if err != nil {
			return nil, err
		}
		left = &cmdBinaryExpr{
			Op:    "==",
			Left:  left,
			Right: right,
		}
	}
	return left, nil
}

func (p *cmdExprParser) parseAdditive() (cmdExpr, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for p.cur.Kind == cmdTokenPlus {
		if err := p.advance(); err != nil {
			return nil, err
		}
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &cmdBinaryExpr{
			Op:    "+",
			Left:  left,
			Right: right,
		}
	}
	return left, nil
}

func (p *cmdExprParser) parseUnary() (cmdExpr, error) {
	if p.cur.Kind == cmdTokenIdent && p.cur.Lit == "typeof" {
		if err := p.advance(); err != nil {
			return nil, err
		}
		v, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &cmdTypeofExpr{Expr: v}, nil
	}
	return p.parsePrimary()
}

func (p *cmdExprParser) parsePrimary() (cmdExpr, error) {
	switch p.cur.Kind {
	case cmdTokenString:
		v := &cmdStringExpr{Value: p.cur.Lit}
		if err := p.advance(); err != nil {
			return nil, err
		}
		return p.parsePostfix(v)
	case cmdTokenNumber:
		v := &cmdNumberExpr{Value: p.cur.Lit}
		if err := p.advance(); err != nil {
			return nil, err
		}
		return p.parsePostfix(v)
	case cmdTokenIdent:
		name := p.cur.Lit
		if err := p.advance(); err != nil {
			return nil, err
		}

		var v cmdExpr
		switch name {
		case "true":
			v = &cmdBoolExpr{Value: true}
		case "false":
			v = &cmdBoolExpr{Value: false}
		case "null":
			v = &cmdNullExpr{}
		default:
			v = &cmdIdentExpr{Name: name}
		}
		return p.parsePostfix(v)
	case cmdTokenLParen:
		if err := p.advance(); err != nil {
			return nil, err
		}
		v, err := p.parseConditional()
		if err != nil {
			return nil, err
		}
		if p.cur.Kind != cmdTokenRParen {
			return nil, fmt.Errorf("expected ')' in expression")
		}
		if err := p.advance(); err != nil {
			return nil, err
		}
		return p.parsePostfix(v)
	default:
		return nil, fmt.Errorf("unexpected token %q", p.cur.Lit)
	}
}

func (p *cmdExprParser) parsePostfix(v cmdExpr) (cmdExpr, error) {
	for {
		switch p.cur.Kind {
		case cmdTokenDot:
			if err := p.advance(); err != nil {
				return nil, err
			}
			if p.cur.Kind != cmdTokenIdent {
				return nil, fmt.Errorf("expected identifier after '.'")
			}
			name := p.cur.Lit
			if err := p.advance(); err != nil {
				return nil, err
			}
			v = &cmdMemberExpr{Target: v, Name: name}
		case cmdTokenLParen:
			args, err := p.parseCallArgs()
			if err != nil {
				return nil, err
			}
			v = &cmdCallExpr{
				Callee: v,
				Args:   args,
			}
		default:
			return v, nil
		}
	}
}

func (p *cmdExprParser) parseCallArgs() ([]cmdExpr, error) {
	// current token is '('
	if err := p.advance(); err != nil {
		return nil, err
	}
	if p.cur.Kind == cmdTokenRParen {
		if err := p.advance(); err != nil {
			return nil, err
		}
		return nil, nil
	}

	args := make([]cmdExpr, 0, 2)
	for {
		arg, err := p.parseConditional()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		switch p.cur.Kind {
		case cmdTokenComma:
			if err := p.advance(); err != nil {
				return nil, err
			}
		case cmdTokenRParen:
			if err := p.advance(); err != nil {
				return nil, err
			}
			return args, nil
		default:
			return nil, fmt.Errorf("expected ',' or ')', got %q", p.cur.Lit)
		}
	}
}
