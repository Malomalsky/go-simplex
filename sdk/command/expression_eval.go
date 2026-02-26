package command

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

func evalCommandExpression(expr string, self any) string {
	payload, err := json.Marshal(self)
	if err != nil {
		panic(fmt.Sprintf("marshal command for expression eval: %v", err))
	}
	var m map[string]any
	if err := json.Unmarshal(payload, &m); err != nil {
		panic(fmt.Sprintf("unmarshal command payload for expression eval: %v", err))
	}

	parser, err := newExprParser(expr, m)
	if err != nil {
		panic(fmt.Sprintf("evaluate command expression: %v", err))
	}
	v, err := parser.parse()
	if err != nil {
		panic(fmt.Sprintf("evaluate command expression: %v", err))
	}
	return jsToString(v)
}

type exprTokenKind int

const (
	tokenEOF exprTokenKind = iota
	tokenIdent
	tokenString
	tokenNumber
	tokenPlus
	tokenQuestion
	tokenColon
	tokenLParen
	tokenRParen
	tokenDot
	tokenComma
	tokenEqEq
	tokenSemicolon
)

type exprToken struct {
	kind exprTokenKind
	lit  string
}

type exprLexer struct {
	src string
	pos int
}

func (l *exprLexer) nextToken() (exprToken, error) {
	l.skipSpaces()
	if l.pos >= len(l.src) {
		return exprToken{kind: tokenEOF}, nil
	}

	ch := l.src[l.pos]
	switch ch {
	case '\'':
		return l.readString()
	case '+':
		l.pos++
		return exprToken{kind: tokenPlus, lit: "+"}, nil
	case '?':
		l.pos++
		return exprToken{kind: tokenQuestion, lit: "?"}, nil
	case ':':
		l.pos++
		return exprToken{kind: tokenColon, lit: ":"}, nil
	case '(':
		l.pos++
		return exprToken{kind: tokenLParen, lit: "("}, nil
	case ')':
		l.pos++
		return exprToken{kind: tokenRParen, lit: ")"}, nil
	case '.':
		l.pos++
		return exprToken{kind: tokenDot, lit: "."}, nil
	case ',':
		l.pos++
		return exprToken{kind: tokenComma, lit: ","}, nil
	case ';':
		l.pos++
		return exprToken{kind: tokenSemicolon, lit: ";"}, nil
	case '=':
		if l.pos+1 < len(l.src) && l.src[l.pos+1] == '=' {
			l.pos += 2
			return exprToken{kind: tokenEqEq, lit: "=="}, nil
		}
		return exprToken{}, fmt.Errorf("unexpected '=' at %d", l.pos)
	}

	if isDigit(ch) {
		return l.readNumber(), nil
	}
	if isIdentStart(ch) {
		return l.readIdent(), nil
	}

	return exprToken{}, fmt.Errorf("unexpected character %q at %d", ch, l.pos)
}

func (l *exprLexer) skipSpaces() {
	for l.pos < len(l.src) {
		switch l.src[l.pos] {
		case ' ', '\t', '\n', '\r':
			l.pos++
		default:
			return
		}
	}
}

func (l *exprLexer) readString() (exprToken, error) {
	// leading '
	l.pos++
	var b strings.Builder
	for l.pos < len(l.src) {
		ch := l.src[l.pos]
		l.pos++
		if ch == '\'' {
			return exprToken{kind: tokenString, lit: b.String()}, nil
		}
		if ch == '\\' {
			if l.pos >= len(l.src) {
				return exprToken{}, fmt.Errorf("unterminated escape at %d", l.pos)
			}
			esc := l.src[l.pos]
			l.pos++
			switch esc {
			case '\\', '\'', '"':
				b.WriteByte(esc)
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			default:
				b.WriteByte(esc)
			}
			continue
		}
		b.WriteByte(ch)
	}
	return exprToken{}, fmt.Errorf("unterminated string literal")
}

func (l *exprLexer) readNumber() exprToken {
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
	return exprToken{kind: tokenNumber, lit: l.src[start:l.pos]}
}

func (l *exprLexer) readIdent() exprToken {
	start := l.pos
	l.pos++
	for l.pos < len(l.src) && isIdentPart(l.src[l.pos]) {
		l.pos++
	}
	return exprToken{kind: tokenIdent, lit: l.src[start:l.pos]}
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

type exprParser struct {
	lexer exprLexer
	self  map[string]any
	cur   exprToken
}

func newExprParser(expr string, self map[string]any) (*exprParser, error) {
	p := &exprParser{
		lexer: exprLexer{src: expr},
		self:  self,
	}
	if err := p.advance(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *exprParser) parse() (any, error) {
	v, err := p.parseConditional()
	if err != nil {
		return nil, err
	}
	if p.cur.kind == tokenSemicolon {
		if err := p.advance(); err != nil {
			return nil, err
		}
	}
	if p.cur.kind != tokenEOF {
		return nil, fmt.Errorf("unexpected token %q", p.cur.lit)
	}
	return v, nil
}

func (p *exprParser) advance() error {
	tok, err := p.lexer.nextToken()
	if err != nil {
		return err
	}
	p.cur = tok
	return nil
}

func (p *exprParser) parseConditional() (any, error) {
	cond, err := p.parseEquality()
	if err != nil {
		return nil, err
	}
	if p.cur.kind != tokenQuestion {
		return cond, nil
	}
	if err := p.advance(); err != nil {
		return nil, err
	}
	whenTrue, err := p.parseConditional()
	if err != nil {
		return nil, err
	}
	if p.cur.kind != tokenColon {
		return nil, fmt.Errorf("expected ':' in conditional expression")
	}
	if err := p.advance(); err != nil {
		return nil, err
	}
	whenFalse, err := p.parseConditional()
	if err != nil {
		return nil, err
	}
	if jsTruthy(cond) {
		return whenTrue, nil
	}
	return whenFalse, nil
}

func (p *exprParser) parseEquality() (any, error) {
	left, err := p.parseAdditive()
	if err != nil {
		return nil, err
	}
	for p.cur.kind == tokenEqEq {
		if err := p.advance(); err != nil {
			return nil, err
		}
		right, err := p.parseAdditive()
		if err != nil {
			return nil, err
		}
		left = jsLooseEqual(left, right)
	}
	return left, nil
}

func (p *exprParser) parseAdditive() (any, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for p.cur.kind == tokenPlus {
		if err := p.advance(); err != nil {
			return nil, err
		}
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = jsAdd(left, right)
	}
	return left, nil
}

func (p *exprParser) parseUnary() (any, error) {
	if p.cur.kind == tokenIdent && p.cur.lit == "typeof" {
		if err := p.advance(); err != nil {
			return nil, err
		}
		v, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return jsTypeOf(v), nil
	}
	return p.parsePrimary()
}

func (p *exprParser) parsePrimary() (any, error) {
	switch p.cur.kind {
	case tokenString:
		v := p.cur.lit
		if err := p.advance(); err != nil {
			return nil, err
		}
		return p.parsePostfix(v)
	case tokenNumber:
		n, err := strconv.ParseFloat(p.cur.lit, 64)
		if err != nil {
			return nil, fmt.Errorf("parse number %q: %w", p.cur.lit, err)
		}
		if err := p.advance(); err != nil {
			return nil, err
		}
		return p.parsePostfix(n)
	case tokenIdent:
		name := p.cur.lit
		if err := p.advance(); err != nil {
			return nil, err
		}
		var v any
		switch name {
		case "self":
			v = p.self
		case "JSON":
			v = jsJSONNamespace{}
		case "true":
			v = true
		case "false":
			v = false
		case "null":
			v = nil
		default:
			return nil, fmt.Errorf("unknown identifier %q", name)
		}
		return p.parsePostfix(v)
	case tokenLParen:
		if err := p.advance(); err != nil {
			return nil, err
		}
		v, err := p.parseConditional()
		if err != nil {
			return nil, err
		}
		if p.cur.kind != tokenRParen {
			return nil, fmt.Errorf("expected ')' in expression")
		}
		if err := p.advance(); err != nil {
			return nil, err
		}
		return p.parsePostfix(v)
	default:
		return nil, fmt.Errorf("unexpected token %q", p.cur.lit)
	}
}

func (p *exprParser) parsePostfix(v any) (any, error) {
	for {
		switch p.cur.kind {
		case tokenDot:
			if err := p.advance(); err != nil {
				return nil, err
			}
			if p.cur.kind != tokenIdent {
				return nil, fmt.Errorf("expected identifier after '.'")
			}
			name := p.cur.lit
			if err := p.advance(); err != nil {
				return nil, err
			}
			if p.cur.kind == tokenLParen {
				args, err := p.parseCallArgs()
				if err != nil {
					return nil, err
				}
				next, err := jsCallMethod(v, name, args)
				if err != nil {
					return nil, err
				}
				v = next
				continue
			}
			next, err := jsGetProp(v, name)
			if err != nil {
				return nil, err
			}
			v = next
		case tokenLParen:
			args, err := p.parseCallArgs()
			if err != nil {
				return nil, err
			}
			next, err := jsCallFunction(v, args)
			if err != nil {
				return nil, err
			}
			v = next
		default:
			return v, nil
		}
	}
}

func (p *exprParser) parseCallArgs() ([]any, error) {
	// current token is '('
	if err := p.advance(); err != nil {
		return nil, err
	}
	if p.cur.kind == tokenRParen {
		if err := p.advance(); err != nil {
			return nil, err
		}
		return nil, nil
	}

	args := make([]any, 0, 2)
	for {
		v, err := p.parseConditional()
		if err != nil {
			return nil, err
		}
		args = append(args, v)

		switch p.cur.kind {
		case tokenComma:
			if err := p.advance(); err != nil {
				return nil, err
			}
		case tokenRParen:
			if err := p.advance(); err != nil {
				return nil, err
			}
			return args, nil
		default:
			return nil, fmt.Errorf("expected ',' or ')', got %q", p.cur.lit)
		}
	}
}

type jsJSONNamespace struct{}

func jsGetProp(v any, name string) (any, error) {
	m, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("property access %q on non-object %T", name, v)
	}
	if field, found := m[name]; found {
		return field, nil
	}
	return nil, nil
}

func jsCallFunction(v any, args []any) (any, error) {
	return nil, fmt.Errorf("unsupported function call on %T with %d args", v, len(args))
}

func jsCallMethod(v any, name string, args []any) (any, error) {
	if _, ok := v.(jsJSONNamespace); ok {
		if name != "stringify" {
			return nil, fmt.Errorf("unsupported JSON method %q", name)
		}
		if len(args) != 1 {
			return nil, fmt.Errorf("JSON.stringify expects 1 arg, got %d", len(args))
		}
		encoded, err := json.Marshal(args[0])
		if err != nil {
			return nil, fmt.Errorf("JSON.stringify: %w", err)
		}
		return string(encoded), nil
	}

	switch name {
	case "toString":
		if len(args) != 0 {
			return nil, fmt.Errorf("toString expects 0 args, got %d", len(args))
		}
		return jsToString(v), nil
	case "join":
		if len(args) > 1 {
			return nil, fmt.Errorf("join expects 0 or 1 args, got %d", len(args))
		}
		items, ok := toSlice(v)
		if !ok {
			return nil, fmt.Errorf("join on non-array %T", v)
		}
		sep := ","
		if len(args) == 1 {
			sep = jsToString(args[0])
		}
		parts := make([]string, len(items))
		for i := range items {
			parts[i] = jsToString(items[i])
		}
		return strings.Join(parts, sep), nil
	default:
		return nil, fmt.Errorf("unsupported method %q on %T", name, v)
	}
}

func toSlice(v any) ([]any, bool) {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return nil, false
	}
	kind := rv.Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		return nil, false
	}
	out := make([]any, rv.Len())
	for i := range out {
		out[i] = rv.Index(i).Interface()
	}
	return out, true
}

func jsAdd(left, right any) any {
	if _, ok := left.(string); ok {
		return jsToString(left) + jsToString(right)
	}
	if _, ok := right.(string); ok {
		return jsToString(left) + jsToString(right)
	}

	if lnum, ok := toNumber(left); ok {
		if rnum, ok := toNumber(right); ok {
			return lnum + rnum
		}
	}
	return jsToString(left) + jsToString(right)
}

func toNumber(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		if err != nil {
			return 0, false
		}
		return f, true
	default:
		return 0, false
	}
}

func jsLooseEqual(left, right any) bool {
	if lnum, ok := toNumber(left); ok {
		if rnum, ok := toNumber(right); ok {
			return !math.IsNaN(lnum) && !math.IsNaN(rnum) && lnum == rnum
		}
	}
	switch l := left.(type) {
	case string:
		r, ok := right.(string)
		return ok && l == r
	case bool:
		r, ok := right.(bool)
		return ok && l == r
	case nil:
		return right == nil
	default:
		return false
	}
}

func jsTruthy(v any) bool {
	switch x := v.(type) {
	case nil:
		return false
	case bool:
		return x
	case string:
		return x != ""
	case float64:
		return x != 0 && !math.IsNaN(x)
	case float32:
		f := float64(x)
		return f != 0 && !math.IsNaN(f)
	case int:
		return x != 0
	case int8:
		return x != 0
	case int16:
		return x != 0
	case int32:
		return x != 0
	case int64:
		return x != 0
	case uint:
		return x != 0
	case uint8:
		return x != 0
	case uint16:
		return x != 0
	case uint32:
		return x != 0
	case uint64:
		return x != 0
	default:
		return true
	}
}

func jsTypeOf(v any) string {
	switch v.(type) {
	case nil:
		// JS quirk: typeof null === "object"
		return "object"
	case bool:
		return "boolean"
	case string:
		return "string"
	case float64, float32, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return "number"
	default:
		return "object"
	}
}

func jsToString(v any) string {
	switch x := v.(type) {
	case nil:
		return "null"
	case string:
		return x
	case bool:
		if x {
			return "true"
		}
		return "false"
	case float64:
		return formatJSNumber(x)
	case float32:
		return formatJSNumber(float64(x))
	case int:
		return strconv.FormatInt(int64(x), 10)
	case int8:
		return strconv.FormatInt(int64(x), 10)
	case int16:
		return strconv.FormatInt(int64(x), 10)
	case int32:
		return strconv.FormatInt(int64(x), 10)
	case int64:
		return strconv.FormatInt(x, 10)
	case uint:
		return strconv.FormatUint(uint64(x), 10)
	case uint8:
		return strconv.FormatUint(uint64(x), 10)
	case uint16:
		return strconv.FormatUint(uint64(x), 10)
	case uint32:
		return strconv.FormatUint(uint64(x), 10)
	case uint64:
		return strconv.FormatUint(x, 10)
	case json.Number:
		return x.String()
	}

	if arr, ok := toSlice(v); ok {
		if len(arr) == 0 {
			return ""
		}
		parts := make([]string, len(arr))
		for i := range arr {
			parts[i] = jsToString(arr[i])
		}
		return strings.Join(parts, ",")
	}

	rv := reflect.ValueOf(v)
	if rv.IsValid() && rv.Kind() == reflect.Map {
		return "[object Object]"
	}

	return fmt.Sprint(v)
}

func formatJSNumber(v float64) string {
	switch {
	case math.IsNaN(v):
		return "NaN"
	case math.IsInf(v, 1):
		return "Infinity"
	case math.IsInf(v, -1):
		return "-Infinity"
	case v == 0:
		// JS prints both 0 and -0 as "0".
		return "0"
	default:
		return strconv.FormatFloat(v, 'f', -1, 64)
	}
}
