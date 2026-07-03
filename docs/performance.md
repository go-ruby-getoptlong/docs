# Performance

`go-ruby-getoptlong/getoptlong` is the pure-Go, CGO-free library that
[`rbgo`](https://github.com/go-embedded-ruby/ruby) binds for Ruby's `getoptlong`.
This page records a **real, library-level** benchmark of that module's Go API
against every reference runtime's own `getoptlong` stdlib, one row per
`GetoptLong` operation. It is part of the ecosystem-wide per-module parity suite,
and **the bar is beating MRI + YJIT**, not just plain MRI.

## What is measured

Two representative `GetoptLong` operations over one **fixed, representative
command line** — six registered options (each a canonical long name plus a short
alias, covering all three of `NO_ARGUMENT`, `REQUIRED_ARGUMENT` and
`OPTIONAL_ARGUMENT`) scanned against a twelve-token argv that exercises every
path: a separated `--opt VALUE`, an `--opt=VALUE`, an optional flag left empty
because an option follows it, a `-dl 3` short bundle (a no-arg short catenated
with a required-arg short that pulls its value from the next word), permuted
operands, and a `--` terminator with two raw words after it:

| Op | What it exercises |
| --- | --- |
| `configure` | construct a `GetoptLong` and build its option tables from all six specs (the constructor path every program pays once) |
| `parse` | construct, then iterate the whole argv to completion — the full lifecycle a real program runs each launch. `GetoptLong` is single-use (it consumes its argument list and terminates), so this is the natural unit of work: long/short/abbreviation matching, `=`-joined and separate values, bundled shorts, optional-argument look-ahead, permuted operands, and the `--` terminator |

The **go-ruby** column drives this pure-Go library through its Go API; every
other column is that interpreter's own stdlib `getoptlong` (a pure-Ruby library).
The Go and Ruby drivers build the **identical** argv, register the **identical**
six options in the same order, and before any timing each op's integer checksum
is verified **byte-identical to MRI** — all five runtimes agree on every op
(`configure`=6, `parse`=507259130). So the comparison is the same observable
operation, apples-to-apples.

- **Host:** Apple M4 Max, macOS (`arm64-darwin`). **Date:** 2026-07-03.
- **Runtimes:** Go 1.26.4; `ruby 4.0.5 +PRISM` (MRI, the oracle, `getoptlong`
  0.2.1) and `ruby --yjit`; `jruby 10.1.0.0` (OpenJDK 25); `truffleruby 34.0.1`
  (GraalVM CE Native).
- **Method:** each process runs 3 untimed warm-up passes then 25 timed passes of
  a fixed inner loop, timed with a monotonic clock; the **best** pass is reported
  as **ns/op**. Interpreter start-up is outside the timed region, so the number
  is the operation's own cost, not `ruby file.rb` process cost. Numbers were
  stable to within a few percent across repeated runs.
- Harness and drivers live in this repo under
  [`benchmarks/`](https://github.com/go-ruby-getoptlong/docs/tree/main/benchmarks)
  (`go/`, `ruby/getoptlong.rb`, `run.sh`). Reproduce: `bash benchmarks/run.sh`.

## Results (ns/op, best of 25)

| Op | go-ruby (pure Go) | MRI | MRI + YJIT | JRuby | TruffleRuby | **go vs YJIT** |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| `configure` | **750** | 9 658 | 7 948 | 10 324 | 4 570 | **10.6× faster** ✅ |
| `parse` | **1 337** | 17 738 | 13 322 | 9 512 | 10 074 | **10.0× faster** ✅ |

## The go-vs-YJIT verdict, per op

**Both operations beat MRI + YJIT, decisively:**

- **`configure` — 10.6× faster than YJIT** (750 ns vs 7 948 ns). Registering six
  options with aliases compiles to a couple of map inserts per name in Go; MRI's
  `getoptlong` builds two Ruby hashes (`@canonical_names`, `@argument_flags`),
  validates each name against a regexp, and appends to an ordered list — all in
  interpreted Ruby, which is where its constructor cost goes.
- **`parse` — 10.0× faster than YJIT** (1 337 ns vs 13 322 ns). Iterating the
  whole argv is a handful of map lookups, slice reslices, and prefix checks per
  token in Go, with no interpreter dispatch at all; MRI runs the same scan as
  Ruby method calls with per-token regexp matching.

Net: **both operations beat MRI + YJIT** (10.0×–10.6×). On `configure`, YJIT is
the faster of the MRI pair and TruffleRuby is the fastest Ruby runtime overall
(4 570 ns) — yet still 6.1× slower than the pure-Go library. On `parse`, JRuby
and TruffleRuby edge ahead of YJIT as their JITs specialise the tight scan loop,
but the pure-Go library is still 7.1× faster than the best of them. Ruby's
`getoptlong` is pure Ruby — not a C extension — so there is no hand-written C to
catch up to here; the compiled pure-Go implementation wins across the board while
producing byte-identical results.

!!! note "Cold-JIT caveat"
    JRuby and TruffleRuby are measured **in-process after 3 warm-up passes**, but
    3 passes is not full JIT steady state for the JVM / GraalVM — their columns
    still carry cold-to-warming-JIT cost and should be read as
    order-of-magnitude, not as fully-warmed peak throughput. Their relatively
    strong `parse` numbers reflect the JIT specialising that loop, yet both remain
    well behind the pure-Go library. The **go-vs-MRI and go-vs-YJIT comparisons
    are the load-bearing ones** — MRI and YJIT reach steady state within the
    warm-up, and both they and the Go driver are timed identically. All numbers
    are real, single-host measurements from the 2026-07-03 run; nothing is
    cherry-picked.
