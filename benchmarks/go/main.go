// Copyright (c) the go-ruby-getoptlong authors
// SPDX-License-Identifier: BSD-3-Clause
//
// Library-level benchmark driver for the pure-Go go-ruby-getoptlong Parser.
// It exercises the same GetoptLong lifecycle as ruby/getoptlong.rb — configuring
// a parser with six options that cover NO_ARGUMENT / REQUIRED_ARGUMENT /
// OPTIONAL_ARGUMENT plus short aliases, then iterating a fixed, representative
// argv to completion — so the ns/op numbers compare this pure-Go library
// primitive against each Ruby runtime's own pure-Ruby getoptlong stdlib.
//
// With CHECK=1 it instead prints one "CHECK\t<label>\t<value>" line per op: an
// integer checksum of the op's result, used to prove the Go output is identical
// to MRI (the oracle) before any timing is trusted.
package main

import (
	"fmt"
	"os"

	"github.com/go-ruby-getoptlong/getoptlong"
)

// argvInput is a fixed, representative command line exercising every scan path:
// a "--opt VALUE" separated required arg, an optional flag left empty because an
// option follows it, an "--opt=VALUE" attached required arg, permuted operands,
// a "-dl 3" bundle (a NO_ARGUMENT short catenated with a REQUIRED_ARGUMENT short
// that pulls its value from the next word), an "--opt=VALUE" optional arg, a
// bare NO_ARGUMENT short, and a "--" terminator followed by two raw words. It is
// byte-for-byte the same argv the Ruby workload builds.
var argvInput = []string{
	"--name", "alice",
	"-v",
	"--output=out.log",
	"input1.txt",
	"-dl", "3",
	"--verbose=high",
	"-h",
	"input2.txt",
	"--", "-x", "raw",
}

// opts registers the same six options, in the same order, as the Ruby driver's
// SPECS: each has a canonical long name, a short alias, and one of the three
// argument flags. Building this slice is outside the timed region on both sides,
// so "configure" measures only the parser's option-table construction.
var opts = []getoptlong.Option{
	{Names: []string{"--help", "-h"}, Flag: getoptlong.NoArgument},
	{Names: []string{"--name", "-n"}, Flag: getoptlong.RequiredArgument},
	{Names: []string{"--verbose", "-v"}, Flag: getoptlong.OptionalArgument},
	{Names: []string{"--output", "-o"}, Flag: getoptlong.RequiredArgument},
	{Names: []string{"--debug", "-d"}, Flag: getoptlong.NoArgument},
	{Names: []string{"--level", "-l"}, Flag: getoptlong.RequiredArgument},
}

// freshArgv returns a private copy of argvInput. GetoptLong consumes its Args
// slice during iteration, so each parse pass needs its own copy — exactly as the
// Ruby driver replaces ARGV with a fresh dup before each pass.
func freshArgv() []string {
	c := make([]string, len(argvInput))
	copy(c, argvInput)
	return c
}

// chkMod bounds the checksum fold so the Go (fixed 64-bit) and Ruby (arbitrary
// precision) accumulators stay bit-identical instead of diverging on overflow.
const chkMod = 1_000_000_007

// checksum folds a full iteration result into one integer identically to the Ruby
// driver: each (name, argument) pair in yield order, then the pair count, then the
// remaining non-option words (Parser.Args) — count and each word's byte length.
func checksum(pairs [][2]string, rest []string) int {
	acc := 0
	for _, p := range pairs {
		acc = (acc*33 + len(p[0])) % chkMod
		acc = (acc*31 + len(p[1])) % chkMod
	}
	acc = (acc*131 + len(pairs)) % chkMod
	acc = (acc*131 + len(rest)) % chkMod
	for _, r := range rest {
		acc = (acc*33 + len(r)) % chkMod
	}
	return acc
}

// opConfigure times parser construction alone: New builds the option tables from
// the pre-made opts slice, with no scanning. Checksum = number of options.
func opConfigure() int {
	p, err := getoptlong.New(freshArgv(), opts...)
	if err != nil {
		panic(err)
	}
	sink = p
	return len(opts)
}

// opParse times the full lifecycle a real GetoptLong program runs each launch:
// construct the parser, then iterate every option to completion (GetoptLong is
// single-use — it consumes its argument list and terminates). Checksum = fold of
// the yielded (name, argument) pairs and the remaining operands.
func opParse() int {
	p, err := getoptlong.New(freshArgv(), opts...)
	if err != nil {
		panic(err)
	}
	p.SetQuiet(true)
	var pairs [][2]string
	if err := p.Each(func(name, argument string) {
		pairs = append(pairs, [2]string{name, argument})
	}); err != nil {
		panic(err)
	}
	return checksum(pairs, p.Args)
}

// ops is the ordered op table shared by the timing and CHECK paths.
var ops = []struct {
	label string
	fn    func() int
}{
	{"configure", opConfigure},
	{"parse", opParse},
}

func main() {
	if os.Getenv("CHECK") != "" {
		for _, o := range ops {
			fmt.Printf("CHECK\t%s\t%d\n", o.label, o.fn())
		}
		return
	}
	const inner = 500
	for _, o := range ops {
		fn := o.fn
		bench(o.label, inner, func() { sink = fn() })
	}
}
