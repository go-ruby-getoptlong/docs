# Roadmap

`go-ruby-getoptlong/getoptlong` is grown **test-first**, each capability differential-tested against
MRI rather than built in isolation. The deterministic, interpreter-independent
slice of Ruby's `GetoptLong` extracted from rbgo's internals is **complete**.

| Stage | What | Status |
| --- | --- | --- |
| Long, short & abbreviated names | Unique-prefix long matching, `AmbiguousOption` on ambiguity, multiple aliases reporting the canonical name. | **Done** |
| Argument forms | `--name=value`, `--name value`, `-nvalue`, `-n value`, bundled `-abc`, with `NoArgument` / `RequiredArgument` / `OptionalArgument` rules. | **Done** |
| Ordering modes | `Permute`, `RequireOrder`, `ReturnInOrder`, plus the `--` terminator. | **Done** |
| Error taxonomy | `InvalidOption`, `MissingArgument`, `NeedlessArgument`, `AmbiguousOption` with MRI-exact messages, `$0` prefix, quiet mode. | **Done** |
| No global state | `Parser`-owned argv and a caller-supplied `io.Writer` instead of global `ARGV` / `$0` / `$stderr`. | **Done** |
| Differential oracle & coverage | All orderings, name/argument forms, terminator, error classes checked byte-for-byte against `ruby` (≥ 4.0); 100% coverage, green across 6 arches and 3 OSes. | **Done** |

## Documented out-of-scope boundaries

These are **deliberate**, recorded so the module's surface is unambiguous:

- **No interpreter.** The library implements the deterministic algorithm; it
  never runs arbitrary Ruby. Anything that needs a live binding or evaluation is
  the consumer's job — that is why `rbgo` binds this module rather than the
  reverse.
- **Reference is reference Ruby (MRI).** Byte-for-byte conformance targets MRI's
  behaviour; differences across MRI releases are matched to the reference used by
  the differential oracle.
- **Standalone & reusable.** The module has no dependency on the Ruby runtime;
  the dependency runs the other way.

See [Usage & API](api.md) for the surface and [Why pure Go](why.md) for the
deterministic/interpreter split.
