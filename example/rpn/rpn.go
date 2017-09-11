package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/hucsmn/peg"
)

const (
	TokenVerb = iota
	TokenNumber
)

var (
	rpnSpaces        = peg.Q1(peg.U("White_Space"))
	rpnOptinalSpaces = peg.Q0(peg.U("White_Space"))
	rpnNewline       = peg.Alt(peg.T("\r\n"), peg.S("\r\n"))
	rpnComment       = peg.Alt(
		peg.Seq(peg.T("#"), peg.UntilB(rpnNewline)),
		peg.Seq(peg.T("("), peg.UntilB(peg.T(")"))))
	rpnVerb   = peg.Q1(peg.U("-White_Space"))
	rpnNumber = peg.Alt(
		peg.Seq(peg.TI("0x"), peg.Q1(peg.R('0', '9', 'a', 'f', 'A', 'F'))),
		peg.Seq(peg.Q01(peg.S("+-")), peg.Q1(peg.R('0', '9'))))

	Builtins = map[string]Primitive{
		"+": Add,
		"-": Sub,
		"*": Mul,
		"/": Div,
		"%": Mod,

		"SWAP": Swap,
		"DUP":  Dup,
		"DROP": Drop,

		".":       Show,
		"EMIT":    Emit,
		"<STACK>": ViewStack,
		"<VOCAB>": ViewVocabulary,
	}

	// Normal cancellation.
	errorMatchCancelled = fmt.Errorf("match cancelled")
)

type (
	State struct {
		stack []int
		vocab map[string]Primitive
	}

	calculator struct {
		*State

		source string
		words  chan peg.Token
		ctx    context.Context
		cancel context.CancelFunc
		pat    peg.Pattern
	}

	Primitive func(*State) error
)

func NewState(vocab map[string]Primitive) *State {
	copied := make(map[string]Primitive)
	for name, prim := range vocab {
		if prim != nil {
			name = strings.ToUpper(name)
			copied[name] = prim
		}
	}
	return &State{vocab: vocab}
}

func (state *State) Calculate(source string) {
	// Send tokens to the channel by registering customed hooks,
	// which are later invoked by peg.Match(pat, source).
	words := make(chan peg.Token, 1)
	ctx, cancel := context.WithCancel(context.Background())
	makeHook := func(ctx context.Context, words chan<- peg.Token, toktype int) func(string, peg.Position) error {
		return func(s string, p peg.Position) error {
			tok := peg.Token{
				Type:     toktype,
				Value:    s,
				Position: p,
			}
			for {
				select {
				case <-ctx.Done():
					return errorMatchCancelled
				case words <- tok:
					return nil
				}
			}
		}
	}

	rpnWord := peg.Alt(
		peg.Trigger(makeHook(ctx, words, TokenNumber), rpnNumber),
		peg.Trigger(makeHook(ctx, words, TokenVerb), rpnVerb),
		rpnComment)
	rpnMain := peg.Seq(rpnOptinalSpaces, peg.J0(rpnWord, rpnSpaces), rpnOptinalSpaces)

	calc := &calculator{
		State:  state,
		source: source,
		words:  words,
		ctx:    ctx,
		cancel: cancel,
		pat:    rpnMain,
	}

	calc.execute()
}

func (calc *calculator) execute() {
	// The lexical scanner routine.
	go func() {
		r, err := peg.Match(calc.pat, calc.source)
		if err != nil {
			if err != errorMatchCancelled {
				fmt.Fprintln(os.Stderr, "error:", err)
			}
		} else if !r.Ok || r.N != len(calc.source) {
			fmt.Fprintln(os.Stderr, "error: invalid program")
		}
		close(calc.words)
	}()

	// Iterate tokens.
loop:
	for tok := range calc.words {
		switch tok.Type {
		case TokenNumber:
			x, err := strconv.ParseInt(tok.Value, 0, 0)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: invalid number %q\n", tok.Value)
				calc.cancel()
				break loop
			}
			calc.Push(int(x))
		case TokenVerb:
			name := tok.Value
			prim := calc.Lookup(name)
			if prim == nil {
				fmt.Fprintf(os.Stderr, "error: undefined verb %q\n", tok.Value)
				calc.cancel()
				break loop
			}
			err := prim(calc.State)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				calc.cancel()
				break loop
			}
		default:
			fmt.Fprintf(os.Stderr, "error: unknown token %q typed %d\n", tok.Value, tok.Type)
			calc.cancel()
			break loop
		}
	}
}

// Instructions.

func (s *State) Push(xs ...int) {
	s.stack = append(s.stack, xs...)
}

func (s *State) Pop(n int) (xs []int, err error) {
	if len(s.stack) < n {
		return nil, fmt.Errorf("stack overflow")
	}

	xs = s.stack[len(s.stack)-n:]
	s.stack = s.stack[:len(s.stack)-n]
	return xs, nil
}

func (s *State) Define(name string, prim Primitive) {
	name = strings.ToUpper(name)
	if s.vocab == nil {
		s.vocab = make(map[string]Primitive)
	}
	s.vocab[name] = prim
}

func (s *State) Lookup(name string) Primitive {
	name = strings.ToUpper(name)
	return s.vocab[name]
}

// Primitives.

func Add(s *State) error {
	xs, err := s.Pop(2)
	if err != nil {
		return err
	}
	s.Push(xs[0] + xs[1])
	return nil
}

func Sub(s *State) error {
	xs, err := s.Pop(2)
	if err != nil {
		return err
	}
	s.Push(xs[0] - xs[1])
	return nil
}

func Mul(s *State) error {
	xs, err := s.Pop(2)
	if err != nil {
		return err
	}
	s.Push(xs[0] * xs[1])
	return nil
}

func Div(s *State) error {
	xs, err := s.Pop(2)
	if err != nil {
		return err
	}
	if xs[1] == 0 {
		return fmt.Errorf("division by zero")
	}
	s.Push(xs[0] / xs[1])
	return nil
}

func Mod(s *State) error {
	xs, err := s.Pop(2)
	if err != nil {
		return err
	}
	if xs[1] == 0 {
		return fmt.Errorf("division by zero")
	}
	s.Push(xs[0] % xs[1])
	return nil
}

func Swap(s *State) error {
	xs, err := s.Pop(2)
	if err != nil {
		return err
	}
	s.Push(xs[1], xs[0])
	return nil
}

func Dup(s *State) error {
	xs, err := s.Pop(1)
	if err != nil {
		return err
	}
	s.Push(xs[0], xs[0])
	return nil
}

func Drop(s *State) error {
	_, err := s.Pop(1)
	if err != nil {
		return err
	}
	return nil
}

func Show(s *State) error {
	xs, err := s.Pop(1)
	if err != nil {
		return err
	}
	fmt.Println(xs[0])
	return nil
}

func Emit(s *State) error {
	xs, err := s.Pop(1)
	if err != nil {
		return err
	}
	fmt.Printf("%c", rune(xs[0]))
	return nil
}

func ViewStack(s *State) error {
	fmt.Printf("stack: %d\n", s.stack)
	return nil
}

func ViewVocabulary(s *State) error {
	strs := make([]string, 0, len(s.vocab))
	for name := range s.vocab {
		strs = append(strs, name)
	}
	sort.Strings(strs)
	fmt.Printf("vocabulary: %q\n", strs)
	return nil
}

// The REPL.

func main() {
	buf := bufio.NewReader(os.Stdin)
	state := NewState(Builtins)
	src := ""
	for {
		fmt.Print(">>> ")

		line, isprefix, err := buf.ReadLine()
		if err != nil {
			break
		}

		src += string(line) + "\n"
		if isprefix {
			continue
		}
		state.Calculate(src)
		src = ""
	}
}
