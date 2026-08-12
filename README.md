# go-icu-collation

A cgo binding to [ICU](https://icu.unicode.org/)'s `ucol_*` collation API for
Go, built for [DumboDB](https://github.com/dolthub/dumbodb)'s MongoDB-compatible
collation.

## Why this exists

MongoDB's collation is ICU. Reproducing it faithfully -- every `strength`,
`caseFirst`, `caseLevel`, `alternate`, `maxVariable`, `normalization`,
`backwards`, and `numericOrdering` option, across locales -- needs ICU itself; no
pure-Go library covers that surface. This module provides exactly the collation
slice of ICU that DumboDB needs, and nothing else.

It **vendors a pinned ICU built statically**, so behavior is fixed to that ICU
version regardless of what the host OS ships. That matters because collation
sort keys become part of DumboDB's on-disk, version-controlled storage format,
which must not shift when the host ICU changes.

It is **independent of** [`go-icu-regex`](https://github.com/dolthub/go-icu-regex):
that module links the *system* ICU (unpinnable) and binds only the regex API,
neither of which serves collation. The two can coexist in one process because
ICU version-suffixes its C symbols. This module follows the same packaging house
style as `go-icu-regex` and [`gozstd`](https://github.com/dolthub/gozstd) but
shares no code with them.

## Status

Early scaffolding. The vendored ICU and the binding land in subsequent changes:

- [ ] Vendor pinned ICU (78.3 / Unicode 17.0) source and a trimmed
      collation-only data blob; build statically via cgo.
- [ ] `ucol_*` binding: `Collator` with `Open`, `SetAttribute`,
      `SetMaxVariable`, `StrColl`, `GetSortKey`, `Version`, `Close`.
- [ ] Unit and concurrency tests.

## Requirements

`CGO_ENABLED=1` and a C/C++ toolchain. Once ICU is vendored, no system ICU is
required -- the module builds its own.

## License

This module's own code is licensed under the Apache License, Version 2.0
([LICENSE](./LICENSE)). It bundles a vendored copy of ICU (under `icu4c/`),
which is under the Unicode License V3 ([icu4c/ICU-LICENSE](./icu4c/ICU-LICENSE)),
retained verbatim per its terms. The Apache license covers only this module's
code, not the vendored ICU. See [NOTICE](./NOTICE) for the full breakdown and
[ACKNOWLEDGEMENTS](./ACKNOWLEDGEMENTS) for attribution.
