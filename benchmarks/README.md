<!-- SPDX-License-Identifier: BSD-3-Clause -->
# `go-ruby-getoptlong` library-level benchmark harness

Reproducible, cross-runtime benchmark of the **pure-Go `go-ruby-getoptlong`
library** against the reference Ruby runtimes (MRI, MRI + YJIT, JRuby,
TruffleRuby). It measures each `GetoptLong` **operation** through the Go API,
isolated from the rbgo interpreter, so the numbers answer: *is the pure-Go
`GetoptLong` as fast as the reference runtime's own pure-Ruby `getoptlong` —
and does it beat MRI + YJIT?*

## Layout

- `go/`                — self-contained Go driver; `go.mod` pins the **published**
  library by pseudo-version (no `replace`).
- `ruby/getoptlong.rb` — the equivalent workload; `ruby/_harness.rb` is the shared
  timer.
- `run.sh`             — runs every available runtime and prints one Markdown table
  per operation (ns/op + ratio vs MRI).

## Run

```sh
bash benchmarks/run.sh
```

Environment knobs: `OUTER` (timed passes, default 25), `WARM` (untimed warm-up
passes, default 3), and `RUBY`/`JRUBY`/`TRUFFLERUBY` to select runtime binaries.

## Method

Each process runs `WARM` untimed passes (to let the JVM/GraalVM JITs warm up),
then `OUTER` timed passes of a fixed inner loop, timed with a monotonic clock;
the **best** pass is reported as **ns/op**. Interpreter start-up is outside the
timed region. The Go driver and the Ruby script build the **identical**
representative argv and register the **identical** six options (short aliases and
all three of `NO_ARGUMENT` / `REQUIRED_ARGUMENT` / `OPTIONAL_ARGUMENT`), and each
op's integer checksum is verified byte-identical across all runtimes
(`CHECK=1 go run .` / `CHECK=1 ruby ruby/getoptlong.rb`) before timing.

`GetoptLong` is single-use — it consumes its argument list and terminates — so
the two ops are `configure` (construct the parser and build its option tables)
and `parse` (construct, then iterate the whole argv to completion, the full
lifecycle a real program runs each launch). Ruby's `GetoptLong` scans the global
`ARGV`; the Go `Parser` scans an explicit slice it owns, so each pass hands it a
fresh copy exactly as the Ruby driver replaces `ARGV` with a fresh dup.

Published, dated results are in [`../docs/performance.md`](../docs/performance.md).
