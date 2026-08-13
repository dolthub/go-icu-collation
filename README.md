# go-icu-collation

A cgo binding to [ICU](https://icu.unicode.org/)'s `ucol_*` collation API for
Go, providing the collation behavior [DumboDB](https://github.com/dolthub/dumbodb)
needs to reproduce MongoDB's.

> **Not a general-purpose library.** This module exists to serve DumboDB and is
> developed solely against DumboDB's needs. Its API, its pinned ICU version, and
> its build choices are driven by that one consumer -- there are no stability
> guarantees, no tagged releases, and no intent to support outside use. If you
> want ICU collation in your own project, bind ICU yourself; please don't depend
> on this.

## Why this exists

MongoDB's collation is ICU. Reproducing it faithfully -- every `strength`,
`caseFirst`, `caseLevel`, `alternate`, `maxVariable`, `normalization`,
`backwards`, and `numericOrdering` option, across locales and their CLDR
tailoring -- needs ICU itself; no pure-Go library covers that surface. This
module exposes exactly the collation slice of ICU that DumboDB uses, and nothing
else.

It **vendors a pinned ICU built statically**, so behavior is fixed to that ICU
version regardless of what the host OS ships. That matters because collation
sort keys become part of DumboDB's on-disk, version-controlled storage format,
which must not shift when the host ICU changes.

It is **independent of** [`go-icu-regex`](https://github.com/dolthub/go-icu-regex):
that module links the *system* ICU (unpinnable) and binds only the regex API,
neither of which serves collation. The two coexist in one process because ICU
version-suffixes its C symbols. This module follows the same packaging house
style as `go-icu-regex` and [`gozstd`](https://github.com/dolthub/gozstd) but
shares no code with them.

## What it provides

The vendored ICU is **78.3 (Unicode 17.0)**, compiled statically via cgo with a
trimmed, collation-only data blob.

`Open(locale)` returns a `*Collator` initialized to that locale's tailoring,
which you configure once and then use concurrently:

- `SetAttribute` / `GetAttribute` -- set or read a collation attribute
  (strength, caseFirst, caseLevel, alternate, numeric, normalization, French/
  backwards). `GetAttribute` returns the *resolved* value, including whatever the
  locale's tailoring implies when you never set it.
- `SetMaxVariable` -- the highest group treated as variable under shifted
  alternate handling.
- `Compare` / `Equal` -- order or compare two strings under the collation.
- `SortKey` -- the collation sort key; byte order matches collation order and
  equal strings share a key.
- `Close` -- release the collator.

A configured `Collator` is immutable, so `Compare`, `Equal`, and `SortKey` are
safe for concurrent use from many goroutines. Package-level `Version()` reports
the linked ICU version, and `SortKey(locale, s)` is a one-shot convenience for
callers that do not hold a `Collator`.

## Requirements

`CGO_ENABLED=1` and a C/C++ toolchain. No system ICU is required -- the module
builds its own vendored copy. (On Linux the C++ runtime is linked explicitly; on
macOS the toolchain supplies it.)

## License

This module's own code is licensed under the Apache License, Version 2.0
([LICENSE](./LICENSE)). It bundles a vendored copy of ICU (under `icu4c/`),
which is under the Unicode License V3 ([icu4c/ICU-LICENSE](./icu4c/ICU-LICENSE)),
retained verbatim per its terms. The Apache license covers only this module's
code, not the vendored ICU. See [NOTICE](./NOTICE) for the full breakdown and
[ACKNOWLEDGEMENTS](./ACKNOWLEDGEMENTS) for attribution.
