# Performance

`go-ruby-getoptlong/getoptlong` is the pure-Go library that
[`rbgo`](https://github.com/go-embedded-ruby/ruby) binds for Ruby's `getoptlong`. This
page records a **comparative benchmark** of that module against the reference
Ruby runtimes, part of the ecosystem-wide per-module parity suite.

## What is measured

The **same** Ruby script — the same workload — scanning a representative argument list (long, short, bundled, `=`-joined) through `GetoptLong#each` — is run under every runtime. `rbgo`'s
number reflects **this pure-Go library doing the work**; every other column is
that interpreter's own `getoptlong` implementation. So the comparison is the
**Ruby-visible operation**, apples-to-apples across interpreters. The script
prints a deterministic checksum and its output is checked **byte-identical to
MRI** before timing.

- **Method:** best-of-N wall time (best, not mean, to suppress scheduler noise);
  single-shot processes, no warm-up beyond the script's own loop.
- **Runtimes:** `ruby` (MRI, the oracle) and `ruby --yjit`; `jruby` (on the JVM);
  `truffleruby` (GraalVM). JRuby and TruffleRuby are timed **cold, single-shot**,
  so they carry JVM / Graal startup on every run.
- The benchmark script and harness live in rbgo's repo under
  [`bench/modules/`](https://github.com/go-embedded-ruby/ruby/tree/main/bench/modules)
  (`getoptlong.rb` + `run.sh`). Reproduce:
  `RBGO=./rbgo TRUFFLE=truffleruby bash bench/modules/run.sh 5`.

!!! note "Cross-runtime numbers pending"
    The full cross-runtime table (rbgo / MRI / YJIT / JRuby / TruffleRuby) for
    this module is produced by the rbgo per-module parity run and is **not yet
    recorded here** — rather than paste fabricated figures, this page documents
    the methodology and points at the harness. The table will be filled in from
    a real measured run, the same way the sibling
    [go-ruby-yaml](https://go-ruby-yaml.github.io/docs/performance/) page reports
    its numbers.

!!! note "Honest framing"
    JRuby and TruffleRuby are timed cold, single-shot, so they carry JVM / Graal
    startup on every run — read them as one-shot `ruby file.rb` costs, the same
    way `rbgo` and MRI are measured, not as steady-state JIT numbers. Rows that
    complete in well under ~200 ms carry the most relative noise; treat their
    ratios as order-of-magnitude.
