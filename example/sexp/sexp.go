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
		peg.Alt(peg.R('a', 'z', 'A', 'Z'), peg.S("!$%&*+,-./:;<=>?@[\\]^_`{|}~")),
		peg.Q0(peg.R('a', 'z', 'A', 'Z', '0', '9'), peg.S("!$%&*+,-./:;<=>?@[\\]^_`{|}~")))
	sexpQuote = peg.T("'")
	sexpLeft  = peg.T("(")
	sexpSep   = peg.Q1(peg.S(" \t\n\r\v\f"))
	sexpRight = peg.T(")")

	sexpRules = map[string]peg.Pattern{
		"atom": peg.Alt(
			// number
			peg.CT(NumberCons, sexpNumber),
			// symbol
			peg.CT(SymbolCons, sexpSymbol),
			// special
			peg.CT(SpecialCons, sexpSpecial)),
		"sexp": peg.Alt(
			// atom
			peg.V("atom"),
			// list
			peg.CC(ListCons,
				peg.Seq(
					sexpLeft, sexpSpaces,
					peg.J0(peg.V("sexp"), sexpSep),
					sexpSpaces, sexpRight)),
			// quoted
			peg.Seq(sexpQuote, peg.CC(QuotedCons, peg.V("sexp")))),
		"incomplete": peg.Alt(
			peg.V("atom"),
			peg.Seq(
				sexpLeft, sexpSpaces,
				peg.J0(peg.V("incomplete"), sexpSep),
				sexpSpaces, peg.Q01(sexpRight)),
			peg.Seq(sexpQuote, peg.V("incomplete"))),
	}
	sexpMain = peg.Let(sexpRules,
		peg.Seq(sexpSpaces, peg.V("sexp"), sexpSpaces))
	sexpIncomplete = peg.Let(sexpRules,
		peg.Seq(sexpSpaces, peg.V("incomplete"), sexpSpaces))
)

// Built-in primitives.
var (
	Builtins = map[string]SExp{
		"+":       Primitive(PrimitiveAdd),
		"-":       Primitive(PrimitiveSub),
		"*":       Primitive(PrimitiveMul),
		"/":       Primitive(PrimitiveDiv),
		"<":       Primitive(PrimitiveLT),
		"<=":      Primitive(PrimitiveLE),
		"==":      Primitive(PrimitiveEQ),
		"!=":      Primitive(PrimitiveNE),
		">=":      Primitive(PrimitiveGE),
		">":       Primitive(PrimitiveGT),
		"not":     Primitive(PrimitiveNot),
		"display": Primitive(PrimitiveDisplay),
		"list":    Primitive(PrimitiveList),
		"nil":     List(nil),
	}
)

// Types.
type (
	SExp interface {
		peg.Capture
		Eval(*Context) (SExp, error)
		Equals(other SExp) bool
		String() string
	}

	Callable interface {
		SExp
		Call(*Context, []SExp) (SExp, error)
	}

	Context struct {
		Scope []map[string]SExp // names are in lower case
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

func (num Number) Equals(other SExp) bool {
	if othernum, ok := other.(Number); ok {
		return float64(num) == float64(othernum)
	}
	return false
}

func (num Number) String() string {
	return fmt.Sprintf("%g", float64(num))
}

// Symbol.

func SymbolCons(lit string, pos peg.Position) (peg.Capture, error) {
	return Symbol(strings.ToLower(lit)), nil
}

func (sym Symbol) IsTerminal() bool {
	return true
}

func (sym Symbol) Eval(ctx *Context) (SExp, error) {
	ret := ctx.Lookup(strings.ToLower(string(sym)))
	if ret == nil {
		return nil, fmt.Errorf("undefined: %s", string(sym))
	}
	return ret, nil
}

func (sym Symbol) Equals(other SExp) bool {
	if othersym, ok := other.(Symbol); ok {
		return strings.ToLower(string(sym)) == strings.ToLower(string(othersym))
	}
	return false
}

func (sym Symbol) String() string {
	return fmt.Sprintf("%s", strings.ToLower(string(sym)))
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

func (b Boolean) Equals(other SExp) bool {
	if otherb, ok := other.(Boolean); ok {
		return b == otherb
	}
	return false
}

func (b Boolean) String() string {
	return fmt.Sprintf("#%t", bool(b))
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

func (prim Primitive) Equals(other SExp) bool {
	if otherprim, ok := other.(Primitive); ok {
		return fmt.Sprintf("%p", prim) == fmt.Sprintf("%p", otherprim)
	}
	return false
}

func (prim Primitive) String() string {
	return fmt.Sprintf("<primitive %p>", prim)
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

func (clr *Closure) Equals(other SExp) bool {
	if otherclr, ok := other.(*Closure); ok {
		return clr == otherclr
	}
	return false
}

func (clr *Closure) String() string {
	return fmt.Sprintf("<closure %p>", clr)
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

func QuotedCons(items []peg.Capture) (peg.Capture, error) {
	if len(items) != 1 {
		return nil, fmt.Errorf("unexpected quoted number: %d", len(items))
	}

	cons := []SExp{Symbol("quote"), items[0].(SExp)}
	return List(cons), nil
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
		case "quote":
			if len(sexps) != 2 {
				return nil, fmt.Errorf("quote syntax requires exactly 1 argument")
			}
			return SyntaxQuote(ctx, sexps[1])
		case "if":
			if len(sexps) != 4 {
				return nil, fmt.Errorf("if syntax requires 3 arguments")
			}
			return SyntaxIf(ctx, sexps[1], sexps[2], sexps[3])
		case "and":
			return SyntaxAnd(ctx, sexps[1:])
		case "or":
			return SyntaxOr(ctx, sexps[1:])
		case "let":
			if len(sexps) != 3 {
				return nil, fmt.Errorf("let syntax requires 2 arguments")
			}
			return SyntaxLet(ctx, sexps[1], sexps[2])
		case "set":
			if len(sexps) != 3 {
				return nil, fmt.Errorf("let syntax requires 2 arguments")
			}
			return SyntaxSet(ctx, sexps[1], sexps[2])
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

func (list List) Equals(other SExp) bool {
	if otherlist, ok := other.(List); ok {
		xs := []SExp(list)
		ys := []SExp(otherlist)
		if len(xs) != len(ys) {
			return false
		}
		for i := range xs {
			if !xs[i].Equals(ys[i]) {
				return false
			}
		}
		return true
	}
	return false
}

func (list List) String() string {
	strs := make([]string, len([]SExp(list)))
	for i := range []SExp(list) {
		strs[i] = fmt.Sprint([]SExp(list)[i])
	}
	return fmt.Sprintf("(%s)", strings.Join(strs, " "))
}

// Predefined Syntax.

func SyntaxQuote(ctx *Context, quoted SExp) (SExp, error) {
	return quoted, nil
}

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

func SyntaxAnd(ctx *Context, args []SExp) (SExp, error) {
	for _, arg := range args {
		evalarg, err := arg.Eval(ctx)
		if err != nil {
			return nil, err
		}
		if b, ok := evalarg.(Boolean); ok {
			if !bool(b) {
				return Boolean(false), nil
			}
		} else {
			return nil, fmt.Errorf("and syntax requires arguments of number type, but got %v", evalarg)
		}
	}
	return Boolean(true), nil
}

func SyntaxOr(ctx *Context, args []SExp) (SExp, error) {
	for _, arg := range args {
		evalarg, err := arg.Eval(ctx)
		if err != nil {
			return nil, err
		}
		if b, ok := evalarg.(Boolean); ok {
			if bool(b) {
				return Boolean(true), nil
			}
		} else {
			return nil, fmt.Errorf("or syntax requires arguments of number type, but got %v", evalarg)
		}
	}
	return Boolean(false), nil
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
					bindings[strings.ToLower(string(name))] = vexpr
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
	val, err := expr.Eval(ctx)
	ctx.Scope = ctx.Scope[:len(ctx.Scope)-1]
	return val, err
}

func SyntaxSet(ctx *Context, lhs, rhs SExp) (SExp, error) {
	var name string
	if sym, ok := lhs.(Symbol); ok {
		name = strings.ToLower(string(sym))
	} else {
		return nil, fmt.Errorf("define syntax requires a symbol in the left hand side, but got %v", lhs)
	}

	val, err := rhs.Eval(ctx)
	if err != nil {
		return nil, err
	}
	top := ctx.Scope[len(ctx.Scope)-1]
	top[name] = val
	return val, nil
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
	scope := make([]map[string]SExp, 2)
	builtins := make(map[string]SExp)
	for k, v := range primitives {
		builtins[strings.ToLower(k)] = v
	}
	scope[0] = builtins
	scope[1] = make(map[string]SExp)
	return &Context{Scope: scope}
}

func (ctx *Context) Lookup(name string) SExp {
	name = strings.ToLower(name)
	for _, top := range ctx.Scope {
		if ret, ok := top[name]; ok {
			return ret
		}
	}
	return nil
}

// Predefined primitive.

func PrimitiveAdd(ctx *Context, args []SExp) (SExp, error) {
	acc := 0.0
	for _, arg := range args {
		if x, ok := arg.(Number); ok {
			acc += float64(x)
			continue
		}
		return nil, fmt.Errorf("'+' requires arguments of number type, but got %v", arg)
	}
	return Number(acc), nil
}

func PrimitiveSub(ctx *Context, args []SExp) (SExp, error) {
	acc := 0.0
	if len(args) > 1 {
		if x, ok := args[0].(Number); ok {
			acc = float64(x)
			args = args[1:]
		} else {
			return nil, fmt.Errorf("'-' requires arguments of number type, but got %v", args[0])
		}
	}

	for _, arg := range args {
		if x, ok := arg.(Number); ok {
			acc -= float64(x)
			continue
		}
		return nil, fmt.Errorf("'-' requires arguments of number type, but got %v", arg)
	}
	return Number(acc), nil
}

func PrimitiveMul(ctx *Context, args []SExp) (SExp, error) {
	acc := 1.0
	for _, arg := range args {
		if x, ok := arg.(Number); ok {
			acc *= float64(x)
			continue
		}
		return nil, fmt.Errorf("'*' requires arguments of number type, but got %v", arg)
	}
	return Number(acc), nil
}

func PrimitiveDiv(ctx *Context, args []SExp) (SExp, error) {
	acc := 1.0
	if len(args) > 1 {
		if x, ok := args[0].(Number); ok {
			acc = float64(x)
			args = args[1:]
		} else {
			return nil, fmt.Errorf("'/' requires arguments of number type, but got %v", args[0])
		}
	}

	for _, arg := range args {
		if x, ok := arg.(Number); ok {
			if float64(x) == 0.0 { // both negative and positive float64 zeroes
				return nil, fmt.Errorf("division by zero")
			}
			acc /= float64(x)
			continue
		}
		return nil, fmt.Errorf("'/' requires arguments of number type, but got %v", arg)
	}
	return Number(acc), nil
}

func PrimitiveLT(ctx *Context, args []SExp) (SExp, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("'<' requires two or more arguments, but got %d", len(args))
	}

	var last float64
	if x, ok := args[0].(Number); ok {
		last = float64(x)
		args = args[1:]
	} else {
		return nil, fmt.Errorf("'<' requires arguments of number type, but got %v", args[0])
	}

	for _, arg := range args {
		if x, ok := arg.(Number); ok {
			if !(last < float64(x)) {
				return Boolean(false), nil
			}
			continue
		}
		return nil, fmt.Errorf("'<' requires arguments of number type, but got %v", arg)
	}
	return Boolean(true), nil
}

func PrimitiveLE(ctx *Context, args []SExp) (SExp, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("'<=' requires two or more arguments, but got %d", len(args))
	}

	var last float64
	if x, ok := args[0].(Number); ok {
		last = float64(x)
		args = args[1:]
	} else {
		return nil, fmt.Errorf("'<=' requires arguments of number type, but got %v", args[0])
	}

	for _, arg := range args {
		if x, ok := arg.(Number); ok {
			if !(last <= float64(x)) {
				return Boolean(false), nil
			}
			continue
		}
		return nil, fmt.Errorf("'<=' requires arguments of number type, but got %v", arg)
	}
	return Boolean(true), nil
}

func PrimitiveEQ(ctx *Context, args []SExp) (SExp, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("'==' requires two or more arguments, but got %d", len(args))
	}

	var first = args[0]
	for _, arg := range args[1:] {
		if !first.Equals(arg) {
			return Boolean(false), nil
		}
	}
	return Boolean(true), nil
}

func PrimitiveNE(ctx *Context, args []SExp) (SExp, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("'!=' requires two or more arguments, but got %d", len(args))
	}

	var first = args[0]
	for _, arg := range args[1:] {
		if !first.Equals(arg) {
			return Boolean(true), nil
		}
	}
	return Boolean(false), nil
}

func PrimitiveGE(ctx *Context, args []SExp) (SExp, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("'>=' requires two or more arguments, but got %d", len(args))
	}

	var last float64
	if x, ok := args[0].(Number); ok {
		last = float64(x)
		args = args[1:]
	} else {
		return nil, fmt.Errorf("'>=' requires arguments of number type, but got %v", args[0])
	}

	for _, arg := range args {
		if x, ok := arg.(Number); ok {
			if !(last >= float64(x)) {
				return Boolean(false), nil
			}
			continue
		}
		return nil, fmt.Errorf("'>=' requires arguments of number type, but got %v", arg)
	}
	return Boolean(true), nil
}

func PrimitiveGT(ctx *Context, args []SExp) (SExp, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("'>' requires two or more arguments, but got %d", len(args))
	}

	var last float64
	if x, ok := args[0].(Number); ok {
		last = float64(x)
		args = args[1:]
	} else {
		return nil, fmt.Errorf("'>' requires arguments of number type, but got %v", args[0])
	}

	for _, arg := range args {
		if x, ok := arg.(Number); ok {
			if !(last > float64(x)) {
				return Boolean(false), nil
			}
			continue
		}
		return nil, fmt.Errorf("'>' requires arguments of number type, but got %v", arg)
	}
	return Boolean(true), nil
}

func PrimitiveNot(ctx *Context, args []SExp) (SExp, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("'not' requires exactly one arguments, but got %d", len(args))
	}

	if b, ok := args[0].(Boolean); ok {
		return Boolean(!bool(b)), nil
	}
	return nil, fmt.Errorf("'not' requires argument of boolean type, but got %v", args[0])
}

func PrimitiveList(ctx *Context, args []SExp) (SExp, error) {
	return List(args), nil
}

func PrimitiveDisplay(ctx *Context, args []SExp) (SExp, error) {
	strs := make([]string, len(args))
	for i := range args {
		strs[i] = fmt.Sprint(args[i])
	}
	fmt.Println(strings.Join(strs, " "))
	return List(nil), nil
}

// The read-evaluate-print loop.

func REPL(ctx *Context, expr string) (val SExp, isprefix bool, err error) {
	// parse.
	caps, err := peg.Parse(sexpMain, expr)
	if err != nil {
		// check if expr is incomplete.
		if peg.IsFullMatched(sexpIncomplete, expr) {
			return nil, true, nil
		}
		return nil, false, err
	}
	if len(caps) != 1 {
		return nil, false, fmt.Errorf("multiple captures: %v", caps)
	}
	sexp, ok := caps[0].(SExp)
	if !ok {
		return nil, false, fmt.Errorf("capture %v is not SExp", caps[0])
	}

	// evaluate.
	val, err = sexp.Eval(ctx)
	return val, false, err
}

func main() {
	buf := bufio.NewReader(os.Stdin)
	ctx := NewContext(Builtins)
	src := ""
	for {
		if src == "" {
			fmt.Print(">>> ")
		} else {
			fmt.Print("... ")
		}
		line, isprefix, err := buf.ReadLine()
		if err != nil {
			break
		}

		src += string(line) + "\n"
		if isprefix {
			continue
		}
		val, isprefix, err := REPL(ctx, src)
		if err != nil {
			src = ""
			fmt.Fprintln(os.Stderr, "error:", err)
			continue
		}
		if isprefix {
			continue
		}
		fmt.Println(val)
		src = ""
	}
}
