# peg

[![Go Doc](https://godoc.org/github.com/hucsmn/peg?status.png)](https://godoc.org/github.com/hucsmn/peg)
[![Build Status](https://travis-ci.org/hucsmn/peg.svg?branch=master)](https://travis-ci.org/hucsmn/peg)
[![Go Report Card](https://goreportcard.com/badge/github.com/hucsmn/peg)](https://goreportcard.com/report/github.com/hucsmn/peg)

This package implements the Parsing Expression Grammars (PEGs),
a powerful tool for pattern matching and writing top-down parsers.
PEGs were designed to focus on expressing the match or parsing progress,
rather than to describe what text should be matched as regexps do.
The package was strongly influenced by [LPeg](http://www.inf.puc-rio.br/~roberto/lpeg/) for lua.
Take a look at it for further readings.

# Overlook

There were four methods for PEGs pattern matching:

```
MatchedPrefix(pat, text) (prefix, ok)
IsFullMatched(pat, text) ok
Parse(pat, text) (captures, err)
Match(pat, text) (result, err)
```

The most general one is `config.Match(pat, text)`, which returns a `*Result`
typed match result and an error if any error occured.

The config tells the max recursion level, the max repeatition times and
whether grouping or capturing is enabled. The default config enables
both grouping and capturing, while limits for recursion and repeat are
setup to DefaultCallstackLimit and DefaultRepeatLimit.

The result of `config.Match(pat, text)` contains:
whether pattern was matched, how many bytes were matched,
the saved groups and the parser captures.
Saved groups are text pieces captured with an optional name.
Parser captures are parse trees or user defined structures constructed
during the parsing process.

Note that, both `MatchedPrefix` and `IsFullMatched` disables capturing.
That is, the side effects of user defined constructors won't be triggered.

# Categories of patterns

Basic patterns, which matches a single rune or a piece of text,
are listed below:

```
T(text), TI(insensitivetext), TS(text, ...), TSI(insensitivetext, ...)
Dot, S(runes), NS(excluderunes), R(low, high, ...), NR(low, high, ...)
U(unicoderangename)
```

Patterns are combined by sequence or alternation:

```
Seq(pat, ...), Alt(pat, ...)
```

Predicators test if pattern would be matched, but consume no text:

```
True, False, SOL, EOL, EOF
B(text), Test(pat), Not(pat), And(pat...), Or(pat...)
When(cond, pat), If(cond, yes, no), Switch(cond, pat, ..., [otherwise])
```

Available pattern qualifiers and repeatitions are:

```
Skip(n), Until(pat), UntilB(pat)
Q0(pat), Q1(pat), Qn(atleast, pat)
Q01(pat), Q0n(atmost, pat), Qnn(exact, pat), Qmn(from, to, pat)
```

Pattern where item separated by sep could be expressed using:

```
J0(item, sep), J1(item, sep), Jn(atleast, item, sep)
J0n(atmost, item, sep), Jnn(exact, item, sep), Jmn(from, to, item, sep)
```

Functionalities for groups, references, triggers and injectors:

```
G(pat), NG(groupname, pat)
Ref(groupname), RefB(groupname)
Trigger(hook, pat), Inject(injector, pat)
Check(checker, pat), Trunc(maxrune, pat)
```

Functionalities for grammars and parsing captures:

```
Let(scope, pat), V(varname), CV(varname), CK(tokentype, pat)
CC(nontermcons, pat), CT(termcons, pat)
```

# Common mistakes

## Greedy qualifiers

The qualifiers are designed to be greedy. Thus, considering the pattern
`Seq(Q0(A), B)`, text supposed to be matched by `B` could be swallowed
ahead of time by the preceding `A`, which is usually unexpected.
It is recommended to wrap `A` with an additional assertion to avoid this.

For example, `Seq(Q0(R('0', '9')), S("02468"), T(" is even"))` is incorrect,
because the greedy `Q0(R('0', '9'))` would consume the last digit, thus the
following `S("02468")` would always dismatch. To make everything right,
`Q0(R('0', '9'))` should be replaced by a pattern like
`Q0(Seq(R('0', '9'), Test(R('0', '9'))))` (assert one digit follow it),
which won't consume the last digit.

## Unreachable branches

Branch of `Seq` or `Alt` could be unreachable, considering that Seq searches
the first dismatch in the sequence, while Alt searches the first match in the
choices. Thus, a pattern like `Alt(T("match"), T("match more"))` would get an
unexpected match result, becuase longer patterns are not in prior order.

## Infinite loops

Any pattern which could macth an empty string should not be nested inside
qualifiers like `Q0`, `Q1`, `Qn`, for this would cause infinite loops.

For example, `Q1(True)` or `Q0(Q0(T("not empty")))` would loop until
`config.RepeatLimit` is reached.

## Left recursion

PEG parsers are top-down, that is, the grammar rules would be expanded
immediately, thus a left recursion never terminates until
`config.CallstackLimit` is reached.

For example, `Let(map[string]Pattern{"var": Seq(T("A"), V("var"))}, V("var"))`
terminates, while
`Let(map[string]Pattern{"var": Seq(V("var"), T("A"))}, V("var"))` won't
terminate until `CallstackLimit` is reached.