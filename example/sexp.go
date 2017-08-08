package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/hucsmn/peg"
)

// PEGs.
var (
	sexpSpaces  = peg.Q0(peg.S(" \t\n\r\v\f"))
	sexpSpecial = peg.Seq(
		peg.T("#"),
		peg.TSI("true", "false"))
	sexpNumber = peg.Seq(
		//sign
		peg.Q01(peg.S("+-")),
		// fraction
		peg.Check(
			func(s string) bool { return s != "." },
			peg.Alt(
				peg.Seq(peg.Q0(peg.R('0', '9')), peg.T("."), peg.Q0(peg.R('0', '9'))),
				peg.Q1(peg.R('0', '9')))),
		// exponent
		peg.Q01(peg.Seq(
			peg.TI("e"),
			peg.Q01(peg.S("+-")),
			peg.Q1(peg.R('0', '9')))))
	sexpSymbol = peg.Seq(
		peg.Alt(peg.R('a', 'z', 'A', 'Z'), peg.S("!$%&'*+,-./:;<=>?@[\\]^_`{|}~")),
		peg.Q0(peg.Alt(peg.R('a', 'z', 'A', 'Z', '0', '9'), peg.S("!$%&'*+,-./:;<=>?@[\\]^_`{|}~"))))
	sexpLeft  = peg.T("(")
	sexpSep   = peg.Q1(peg.S(" \t\n\r\v\f"))
	sexpRight = peg.T(")")

	sexpMain = peg.Let(
		map[string]peg.Pattern{
			"sexp": peg.Alt(
				// number
				peg.CT(NumberCons, sexpNumber),
				// symbol
				peg.CT(SymbolCons, sexpSymbol),
				// special
				peg.CT(SpecialCons, sexpSpecial),
				// list
				peg.CC(ListCons,
					peg.Seq(
						sexpLeft, sexpSpaces,
						peg.J0(peg.V("sexp"), sexpSep),
						sexpSpaces, sexpRight))),
		},
		peg.Seq(sexpSpaces, peg.V("sexp"), sexpSpaces))
)

// Primitives.
var (
	Builtins = map[string]SExp{
		"+": Primitive(PrimitiveAdd),
		"-": Primitive(PrimitiveSub),
		"*": Primitive(PrimitiveMul),
		"/": Primitive(PrimitiveDiv),
	}
)

// Types.
type (
	SExp interface {
		peg.Capture
		Eval(*Context) (SExp, error)
	}

	Callable interface {
		SExp
		Call(*Context, []SExp) (SExp, error)
	}

	Context struct {
		Scope []map[string]SExp
	}

	List []SExp

	Symbol string

	Number float64

	Boolean bool

	Primitive func(*Context, []SExp) (SExp, error)

	Closure struct {
		bind []map[string]SExp
		args []string
		body SExp
	}
)

// Number.

func NumberCons(lit string, pos peg.Position) (peg.Capture, error) {
	var sign float64
	if strings.HasPrefix(lit, "+") {
		lit = lit[1:]
		sign = 1.0
	} else if strings.HasPrefix(lit, "-") {
		lit = lit[1:]
		sign = -1.0
	} else {
		sign = 1.0
	}

	num, err := strconv.ParseFloat(lit, 64)
	if err != nil {
		return nil, err
	}
	num *= sign
	return Number(num), nil
}

func (num Number) IsTerminal() bool {
	return true
}

func (num Number) Eval(ctx *Context) (SExp, error) {
	return num, nil
}

// Symbol.

func SymbolCons(lit string, pos peg.Position) (peg.Capture, error) {
	return Symbol(lit), nil
}

func (sym Symbol) IsTerminal() bool {
	return true
}

func (sym Symbol) Eval(ctx *Context) (SExp, error) {
	ret := ctx.Lookup(string(sym))
	if ret == nil {
		return nil, fmt.Errorf("undefined: %s", string(sym))
	}
	return ret, nil
}

// Boolean.

func SpecialCons(lit string, pos peg.Position) (peg.Capture, error) {
	switch strings.ToLower(lit) {
	case "#true":
		return Boolean(true), nil
	case "#false":
		return Boolean(false), nil
	}
	return nil, fmt.Errorf("unknown special literal %q", lit)
}

func (b Boolean) IsTerminal() bool {
	return true
}

func (b Boolean) Eval(ctx *Context) (SExp, error) {
	return b, nil
}

// Primitive.

func (prim Primitive) IsTerminal() bool {
	return true
}

func (prim Primitive) Eval(ctx *Context) (SExp, error) {
	return prim, nil
}

func (prim Primitive) Call(ctx *Context, args []SExp) (SExp, error) {
	return prim(ctx, args)
}

// Closure.

func (clr *Closure) IsTerminal() bool {
	return true
}

func (clr *Closure) Eval(ctx *Context) (SExp, error) {
	return clr, nil
}

func (clr *Closure) Call(ctx *Context, args []SExp) (SExp, error) {
	if len(clr.args) != len(args) {
		return nil, fmt.Errorf("closure %p requies %d arguments, but got %d",
			clr, len(clr.args), len(args))
	}

	// build namespace from clr.bind and args.
	backup := ctx.Scope
	ctx.Scope = make([]map[string]SExp, len(clr.bind)+1)
	copy(ctx.Scope, clr.bind)
	top := make(map[string]SExp)
	for i := range args {
		top[clr.args[i]] = args[i]
	}
	ctx.Scope[len(ctx.Scope)-1] = top

	// invoke inner SExp and recover namespace.
	ret, err := clr.body.Eval(ctx)
	ctx.Scope = backup
	return ret, err
}

// List.

func ListCons(items []peg.Capture) (peg.Capture, error) {
	sexps := make([]SExp, len(items))
	var ok bool
	for i := range items {
		sexps[i], ok = items[i].(SExp)
		if !ok {
			return nil, fmt.Errorf("unexpected capture: %#v", items[i])
		}
	}
	return List(sexps), nil
}

func (list List) IsTerminal() bool {
	return false
}

func (list List) Eval(ctx *Context) (SExp, error) {
	sexps := []SExp(list)

	// nil.
	if len(sexps) == 0 {
		return list, nil
	}

	// predefined syntax.
	if sym, ok := sexps[0].(Symbol); ok {
		switch strings.ToLower(string(sym)) {
		case "if":
			if len(sexps) != 4 {
				return nil, fmt.Errorf("if syntax requires 3 arguments")
			}
			return SyntaxIf(ctx, sexps[1], sexps[2], sexps[3])
		case "let":
			if len(sexps) != 3 {
				return nil, fmt.Errorf("let syntax requires 2 arguments")
			}
			return SyntaxLet(ctx, sexps[1], sexps[2])
		case "lambda":
			if len(sexps) != 3 {
				return nil, fmt.Errorf("lambda syntax requires 2 arguments")
			}
			return SyntaxLambda(ctx, sexps[1], sexps[2])
		}
	}

	// simple function call.
	evals := make([]SExp, len(sexps))
	for i := range sexps {
		var err error
		evals[i], err = sexps[i].Eval(ctx)
		if err != nil {
			return nil, err
		}
	}

	fn, ok := evals[0].(Callable)
	if !ok {
		return nil, fmt.Errorf("non-callable: %v", evals[0])
	}
	return fn.Call(ctx, evals[1:])
}

// Predefined Syntax.

func SyntaxIf(ctx *Context, cond, yes, no SExp) (SExp, error) {
	evalcond, err := cond.Eval(ctx)
	if err != nil {
		return nil, err
	}
	if b, ok := evalcond.(Boolean); ok {
		var ret SExp
		if bool(b) {
			ret, err = yes.Eval(ctx)
		} else {
			ret, err = no.Eval(ctx)
		}
		return ret, err
	}
	return nil, fmt.Errorf("if syntax requires condition to be a boolean, but got %v", evalcond)
}

func SyntaxLet(ctx *Context, bind, expr SExp) (SExp, error) {
	// parse bindings.
	bindings := make(map[string]SExp)
	if list, ok := bind.(List); ok {
		for _, i := range []SExp(list) {
			if pair, ok := i.(List); ok && len([]SExp(pair)) == 2 {
				ksym := []SExp(pair)[0]
				vexpr := []SExp(pair)[1]
				if name, ok := ksym.(Symbol); ok {
					bindings[string(name)] = vexpr
					continue
				}
			}
			return nil, fmt.Errorf("bad let syntax binding-pair %v", i)
		}
	} else {
		return nil, fmt.Errorf("let syntax requires binding-pairs, but got %v", bind)
	}

	// evaluate bindings and build namespace.
	top := make(map[string]SExp)
	ctx.Scope = append(ctx.Scope, top)
	for name, _ := range bindings {
		// initialize to nil.
		top[name] = List(nil)
	}
	for name, vexpr := range bindings {
		value, err := vexpr.Eval(ctx)
		if err != nil {
			return nil, err
		}
		top[name] = value
	}

	// evaluate inner SExpr and recover namespace.
	ret, err := expr.Eval(ctx)
	ctx.Scope = ctx.Scope[:len(ctx.Scope)-1]
	return ret, err
}

func SyntaxLambda(ctx *Context, args, expr SExp) (SExp, error) {
	clr := &Closure{}

	// parse arguments.
	if list, ok := args.(List); ok {
		clr.args = make([]string, len([]SExp(list)))
		for i := range []SExp(list) {
			arg := []SExp(list)[i]
			if sym, ok := arg.(Symbol); ok {
				clr.args[i] = string(sym)
				continue
			}
			return nil, fmt.Errorf("bad lambda syntax argument %v", arg)
		}
	} else {
		return nil, fmt.Errorf("lambda syntax requires arguments list, but got %v", args)
	}

	// snapshot namespace.
	clr.bind = make([]map[string]SExp, len(ctx.Scope))
	copy(clr.bind, ctx.Scope)

	// build closure.
	clr.body = expr
	return clr, nil
}

// Context.

func NewContext(primitives map[string]SExp) *Context {
	scope := make([]map[string]SExp, 1)
	top := make(map[string]SExp)
	for k, v := range primitives {
		top[k] = v
	}
	scope[0] = top
	return &Context{Scope: scope}
}

func (ctx *Context) Lookup(name string) SExp {
	for _, top := range ctx.Scope {
		if ret, ok := top[name]; ok {
			return ret
		}
	}
	return nil
}

// Predefined primitive.

func PrimitiveAdd(ctx *Context, args []SExp) (SExp, error) {
	var acc float64
	for _, arg := range args {
		if x, ok := arg.(Number); ok {
			acc += float64(x)
			continue
		}
		return nil, fmt.Errorf("add requires arguments of number type, but got %v", arg)
	}
	return Number(acc), nil
}

func PrimitiveSub(ctx *Context, args []SExp) (SExp, error) {
	var acc float64
	for _, arg := range args {
		if x, ok := arg.(Number); ok {
			acc -= float64(x)
			continue
		}
		return nil, fmt.Errorf("sub requires arguments of number type, but got %v", arg)
	}
	return Number(acc), nil
}

func PrimitiveMul(ctx *Context, args []SExp) (SExp, error) {
	var acc float64
	for _, arg := range args {
		if x, ok := arg.(Number); ok {
			acc *= float64(x)
			continue
		}
		return nil, fmt.Errorf("mul requires arguments of number type, but got %v", arg)
	}
	return Number(acc), nil
}

func PrimitiveDiv(ctx *Context, args []SExp) (SExp, error) {
	var acc float64
	for _, arg := range args {
		if x, ok := arg.(Number); ok {
			if float64(x) == 0.0 { // both negative and positive float64 zeroes
				return nil, fmt.Errorf("division by zero")
			}
			acc /= float64(x)
			continue
		}
		return nil, fmt.Errorf("div requires arguments of number type, but got %v", arg)
	}
	return Number(acc), nil
}

// The evaluate function.
func Eval(expr string) (SExp, error) {
	// parse.
	caps, err := peg.Parse(sexpMain, expr)
	if err != nil {
		return nil, err
	}
	if len(caps) != 1 {
		return nil, fmt.Errorf("multiple captures: %v", caps)
	}
	sexp, ok := caps[0].(SExp)
	if !ok {
		return nil, fmt.Errorf("capture %v is not SExp", caps[0])
	}

	// evaluate.
	ctx := NewContext(Builtins)
	return sexp.Eval(ctx)
}

func main() {
	buf := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("sexp> ")
		line, _, err := buf.ReadLine()
		if err != nil {
			break
		}
		fmt.Println(Eval(string(line)))
	}
}
