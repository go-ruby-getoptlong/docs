# Usage & API

The public API lives at the module root (`github.com/go-ruby-getoptlong/getoptlong`). It is **Ruby-shaped but Go-idiomatic**: `GetNext` / `Each` mirror `GetoptLong#get` / `#each`, while the surface follows Go conventions — a `Parser`-owned argv, an explicit `error`, a caller-supplied `io.Writer`, no global `ARGV`.

!!! success "Status: implemented"
    The library is built and importable as
    `github.com/go-ruby-getoptlong/getoptlong`, bound into `rbgo` as a native
    module; see [Roadmap](roadmap.md).

## Install

```sh
go get github.com/go-ruby-getoptlong/getoptlong
```

## Worked example

```go
p, err := getoptlong.New(os.Args[1:],
    getoptlong.Option{Names: []string{"--number", "-n"}, Flag: getoptlong.RequiredArgument},
    getoptlong.Option{Names: []string{"--verbose", "-v"}, Flag: getoptlong.OptionalArgument},
    getoptlong.Option{Names: []string{"--help", "-h"}, Flag: getoptlong.NoArgument},
)
if err != nil {
    panic(err) // a *SpecError: a malformed option specification
}
p.ProgName = "fib"        // mirrors Ruby's $0 in error messages
p.ErrorWriter = os.Stderr // mirrors Ruby's $stderr

err = p.Each(func(name, argument string) {
    fmt.Printf("%-10s %q\n", name, argument)
})
if err != nil { // *getoptlong.Error: InvalidOption / MissingArgument / ...
    os.Exit(1)
}
fmt.Println("operands:", p.Args) // the remaining, non-option arguments
```

```
$ fib --number 6 -v -- extra
--number   "6"
--verbose  ""
operands: [extra]
```

## Ordering

The three ordering modes mirror Ruby's `GetoptLong` exactly (set with
`SetOrdering`). Given `Foo --zzz Bar --xxx Baz Bat` with `--xxx` taking a
required argument and `--zzz` taking none:

| Ordering        | Options yielded                                  | `p.Args` after |
| --------------- | ------------------------------------------------ | -------------- |
| `Permute`       | `("--zzz","")`, `("--xxx","Baz")`                | `[Foo Bat]`    |
| `RequireOrder`  | *(none — first word is an operand)*              | `[Foo --zzz Bar --xxx Baz Bat]` |
| `ReturnInOrder` | `("","Foo")`, `("--zzz","")`, `("--xxx","Baz")`, `("","Bat")` | `[]` |

Unlike Ruby, this package does **not** consult `POSIXLY_CORRECT`; pass
`RequireOrder` explicitly for POSIX behaviour.

## Surface

```go
func New(args []string, options ...Option) (*Parser, error) // PERMUTE ordering
func NewParser() *Parser

type ArgumentFlag int
const (NoArgument; RequiredArgument; OptionalArgument)
type Ordering int
const (RequireOrder; Permute; ReturnInOrder)
type Option struct { Names []string; Flag ArgumentFlag }

type Parser struct {
    Args        []string  // working argv; leftover operands after scanning
    ProgName    string    // error-message prefix (Ruby's $0)
    ErrorWriter io.Writer // error sink (Ruby's $stderr); nil discards
    // ...
}

func (p *Parser) GetNext() (name, argument string, ok bool, err error) // GetoptLong#get
func (p *Parser) Each(fn func(name, argument string)) error            // GetoptLong#each
func (p *Parser) SetOptions(options ...Option) error
func (p *Parser) SetOrdering(o Ordering) error; func (p *Parser) Ordering() Ordering
func (p *Parser) SetQuiet(q bool); func (p *Parser) Quiet() bool
func (p *Parser) Err() *Error; func (p *Parser) ErrorMessage() string
func (p *Parser) Terminated() bool

type Error struct { Kind ErrorKind; Message string } // GetoptLong::Error
type ErrorKind int                                   // InvalidOption / MissingArgument /
                                                     // NeedlessArgument / AmbiguousOption
type SpecError struct { Message string }             // ArgumentError from New / SetOptions
```

`GetNext` returns `("", "", false, nil)` when there are no more options; under
`ReturnInOrder` an operand is returned as `("", word, true, nil)`. On a parse
error it returns `("", "", false, *Error)`, records the error
(`Err` / `ErrorMessage`), and subsequent calls yield nothing.

## MRI conformance

A **differential oracle** runs every scenario — all three orderings, long / short
/ abbreviated names, `=`-joined / separate / bundled arguments, the `--`
terminator, and each error class — through both this package and the system
`ruby`, comparing the option stream, leftover argv, and error class + message
**byte-for-byte**. It is gated on `RUBY_VERSION >= "4.0"` and skips itself where
`ruby` is absent so the cross-arch lanes still validate the library.
