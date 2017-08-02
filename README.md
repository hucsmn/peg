# peg

Parsing Expression Grammars (PEG) is a powerful tool for pattern matching,
text extraction and parse tree building. The PEG text matching is greedy,
that is, the qualified patterns try to match as more bytes as it could.
The PEG parsers are top-down parsers similar to LL parsers. Thus, left
recursion should be particularly avoided when writing grammar rules. The
package design was strongly influenced by lua's [LPeg](http://www.inf.puc-rio.br/~roberto/lpeg/).


# Overlook of methods

There are four methods for PEG pattern matching, text extracting and
parse tree building:

```
MatchedPrefix(pat, text) (prefix, ok)
IsFullMatched(pat, text) ok
Match(pat, text) (result, err)
ConfiguredMatch(config, pat, text) (result, err)
```

The configuration `Config` of pattern matching determines max recursion/loop
times and whether some functionality is enabled/disabled.
The result of `Result` type contains: is matched, count of bytes matched,
saved groups `Groups` and `NamedGroups` and the parser captures of
`[]Capture` type.
Saved groups are text pieces captured with an optional name.
Parse captures are parse trees or user defined structures constructed during
parsing process.

# Overlook of patterns
There are several basic patterns, which matches a single rune or a piece of text:
```
T(text), TI(insensitivetext), TS(text, ...), TSI(insensitivetext, ...)
Dot, S(runes), NS(excluderunes), R(low, high, ...), NR(low, high, ...)
U(unicoderangename)
```
Patterns are combined by sequence or alternation:
```
Seq(pat, ...), Alt(pat, ...)
```
There are some predicators which test if pattern matches but consumes no text:
```
True, False, SOL, EOL, EOF
B(text), Test(pat), Not(pat), And(pat...), Or(pat...)
When(cond, pat), If(cond, yes, no), Switch(cond, pat, ..., [otherwise])
```
Available qualifiers for patterns are:
```
Q0(pat), Q1(pat), Qn(atleast, pat)
Q01(pat), Q0n(atmost, pat), Qnn(exact, pat), Qmn(from, to, pat)
```
Join helpers are:
```
J0(item, sep), Jn(atleast, item, sep)
J0n(atmost, item, sep), Jnn(exact, item, sep), Jmn(from, to, item, sep)
```
Supports for groups, references, triggers and injectors:
```
G(pat), NG(groupname, pat)
Ref(groupname), RefB(groupname)
Trigger(hook, pat), Save(pointer, pat), Send(channel, pat), SendToken(channel, tokentype, pat)
Inject(injector, pat), Check(checker, pat), Trunc(maxrune, pat)
```
Supports for parser capturing:
```
Let(scope, pat), V(varname), CV(varname), CK(tokentype, pat)
CC(nontermcons, pat), CT(termcons, pat)
```
# Common mistakes

## Unreachable branches

Branch of `Seq` or `Alt` could be unreachable, considering that Seq searches the
first dismatch in the sequence, while Alt searches the first match in the
choices. Thus, both `Seq(False, unreachable)` and `Alt(True, unreachable)` could
just be a mistake. The cases like `Alt(T("match"), T("match more"))` is common
mistakes where the pattern matching more text is not in a prior order.

## Infinite loops

Any pattern that macthes empty string should not be directly nested inside
a qualifier like `Q0`, `Q1`, `Qn`. It may result in an infinite loop. For example,
`Q1(True)` or `Q0(Q0(T("not empty")))` would loop until `LoopLimit` is reached.

## Left recursion

PEG parsers are top-down, that is, the context-free grammar rules would be
expanded immediately, thus a left recursion would never terminate.
For example, `Let(map[string]Pattern{"var": Seq(T("A"), V("var"))}, V("var"))`
terminates, while
`Let(map[string]Pattern{"var": Seq(V("var"), T("A"))}, V("var"))` won't
terminate until `CallstackLimit` is reached.