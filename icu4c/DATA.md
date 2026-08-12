# Vendored ICU data (`icudt78l.dat`)

`icudt78l.dat` is a **collation-only** trim of ICU 78.3's data, ~3.8 MB (the full
`icudt78l.dat` is ~32 MB). It carries only what collation needs; every other data
category (currency, timezone, region, language, unit, formatting, break
iterators, transliteration, converters) is removed.

`TestMongoLocaleFingerprints` guards the trim: it pins the sort keys of a set of
tailoring probes for every locale MongoDB supports, and fails if a trim ever
changes a locale's collation. The checked-in fingerprints were generated from the
*full* data, so a passing test proves the trim dropped nothing collation-relevant.

## Items kept

- `coll/*` -- the collation tree: `root`, `ucadata.icu` (the UCA table), every
  locale tailoring, and `res_index`.
- `*.nrm` -- normalization tables.
- `pool.res` -- the shared string pool the resource bundles reference.
- `res_index.res`, `root.res` -- top-level locale index and root bundle.

## Regenerating

Trimming is done with ICU's `pkgdata` (the runtime data packager), not `icupkg`.
On this ICU release `icupkg` cannot round-trip the full package -- its
`STRING_STORE_SIZE` (100000 in `tools/toolutil/package.h`) overflows on the
~4300-item table and it emits packages ICU cannot load. `pkgdata` has no such
limit and produces the runtime format directly.

From a built ICU 78.3 source tree (`icu4c-78.3-sources.tgz`, configured and
`make`d so `bin/pkgdata` and `data/out/build/icudt78l/` exist):

    # keep the collation tree plus its dependencies
    grep -E '^coll/|\.nrm$|^pool\.res$|^res_index\.res$|^root\.res$' \
        data/out/tmp/icudata.lst | sort > coll.lst

    # package just those into a single common-data blob
    LD_LIBRARY_PATH=./lib ./bin/pkgdata \
        -O data/icupkg.inc -q -c \
        -s data/out/build/icudt78l -d <outdir> \
        -e icudt78 -T <tmpdir> -p icudt78l -m common coll.lst

The resulting `<outdir>/icudt78l.dat` replaces this file. Re-run the icu4c tests;
`TestMongoLocaleFingerprints` must still pass.
