# frozen_string_literal: true
# Copyright (c) the go-ruby-getoptlong authors
# SPDX-License-Identifier: BSD-3-Clause
#
# Reference GetoptLong workload, mirroring benchmarks/go/main.go op-for-op over an
# identical, representative argv. Run normally it reports ns/op per op through the
# shared harness; run with CHECK=1 it prints one "CHECK\t<label>\t<value>" line per
# op so the Go output can be proven identical to MRI (the oracle) before any timing
# is trusted.
require "getoptlong"
require_relative "_harness"

# Byte-for-byte the same argv the Go driver builds: a "--opt VALUE" separated
# required arg, an optional flag left empty because an option follows it, an
# "--opt=VALUE" attached required arg, permuted operands, a "-dl 3" bundle (a
# NO_ARGUMENT short catenated with a REQUIRED_ARGUMENT short that pulls its value
# from the next word), an "--opt=VALUE" optional arg, a bare NO_ARGUMENT short,
# and a "--" terminator followed by two raw words.
ARGV_INPUT = [
  "--name", "alice",
  "-v",
  "--output=out.log",
  "input1.txt",
  "-dl", "3",
  "--verbose=high",
  "-h",
  "input2.txt",
  "--", "-x", "raw",
].freeze

# SPECS registers the same six options, in the same order, as the Go driver's
# opts: each has a canonical long name, a short alias, and one of the three
# argument flags. Building this array is outside the timed region, so "configure"
# measures only GetoptLong.new's option-table construction.
SPECS = [
  ["--help",    "-h", GetoptLong::NO_ARGUMENT],
  ["--name",    "-n", GetoptLong::REQUIRED_ARGUMENT],
  ["--verbose", "-v", GetoptLong::OPTIONAL_ARGUMENT],
  ["--output",  "-o", GetoptLong::REQUIRED_ARGUMENT],
  ["--debug",   "-d", GetoptLong::NO_ARGUMENT],
  ["--level",   "-l", GetoptLong::REQUIRED_ARGUMENT],
].freeze

# CHK_MOD bounds the checksum fold so the Ruby (arbitrary precision) and Go
# (fixed 64-bit) accumulators stay bit-identical instead of diverging on overflow.
CHK_MOD = 1_000_000_007

# checksum folds a full iteration result into one integer identically to the Go
# driver: each (name, argument) pair in yield order, then the pair count, then the
# remaining non-option words left in ARGV — count and each word's byte length.
def checksum(pairs, rest)
  acc = 0
  pairs.each do |name, argument|
    acc = (acc * 33 + name.bytesize) % CHK_MOD
    acc = (acc * 31 + argument.bytesize) % CHK_MOD
  end
  acc = (acc * 131 + pairs.length) % CHK_MOD
  acc = (acc * 131 + rest.length) % CHK_MOD
  rest.each { |r| acc = (acc * 33 + r.bytesize) % CHK_MOD }
  acc
end

# configure: construct the parser alone (build the option tables, no scanning).
# Checksum = number of option specs.
def op_configure
  $sink = GetoptLong.new(*SPECS)
  SPECS.length
end

# parse: the full lifecycle a real program runs — construct, then iterate every
# option to completion. GetoptLong is single-use: it consumes ARGV and terminates,
# leaving the remaining operands in ARGV. Checksum = fold of the yielded pairs and
# the remaining operands.
def op_parse
  ARGV.replace(ARGV_INPUT.dup)
  go = GetoptLong.new(*SPECS)
  go.quiet = true
  pairs = []
  go.each { |name, argument| pairs << [name, argument] }
  checksum(pairs, ARGV)
end

OPS = [
  ["configure", method(:op_configure)],
  ["parse",     method(:op_parse)],
].freeze

if ENV["CHECK"] && !ENV["CHECK"].empty?
  OPS.each { |label, m| printf("CHECK\t%s\t%d\n", label, m.call) }
else
  INNER = 500
  OPS.each { |label, m| bench(label, INNER) { m.call } }
end
