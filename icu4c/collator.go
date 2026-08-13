// Copyright 2026 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package icu4c

// #include "unicode/ucol.h"
// #include "unicode/utypes.h"
// #include <stdlib.h>
//
// static UCollator *icu4c_ucol_open(const char *loc, char *err, int errcap) {
//     UErrorCode status = U_ZERO_ERROR;
//     UCollator *c = ucol_open(loc, &status);
//     if (c == NULL || U_FAILURE(status)) {
//         if (err != NULL && errcap > 0) {
//             const char *name = u_errorName(status);
//             int i = 0;
//             for (; name[i] != 0 && i < errcap - 1; i++) err[i] = name[i];
//             err[i] = 0;
//         }
//         if (c != NULL) ucol_close(c);
//         return NULL;
//     }
//     return c;
// }
//
// static int icu4c_ucol_set_attribute(UCollator *c, UColAttribute a, UColAttributeValue v) {
//     UErrorCode status = U_ZERO_ERROR;
//     ucol_setAttribute(c, a, v, &status);
//     return (int)status;
// }
//
// static int icu4c_ucol_get_attribute(UCollator *c, UColAttribute a, int *out) {
//     UErrorCode status = U_ZERO_ERROR;
//     UColAttributeValue v = ucol_getAttribute(c, a, &status);
//     if (U_FAILURE(status)) return (int)status;
//     *out = (int)v;
//     return 0;
// }
//
// static int icu4c_ucol_set_max_variable(UCollator *c, UColReorderCode group) {
//     UErrorCode status = U_ZERO_ERROR;
//     ucol_setMaxVariable(c, group, &status);
//     return (int)status;
// }
//
// static int icu4c_ucol_strcoll(UCollator *c, const char *a, int alen, const char *b, int blen) {
//     UErrorCode status = U_ZERO_ERROR;
//     UCollationResult r = ucol_strcollUTF8(c, a, alen, b, blen, &status);
//     if (U_FAILURE(status)) return -2;
//     return (int)r; // UCOL_LESS(-1) / UCOL_EQUAL(0) / UCOL_GREATER(1)
// }
//
// static int icu4c_ucol_sortkey(UCollator *c, const UChar *s, int slen, uint8_t *out, int cap) {
//     return ucol_getSortKey(c, s, slen, out, cap);
// }
import "C"

import (
	"fmt"
	"unicode/utf16"
	"unsafe"
)

// Attribute names an ICU collator attribute (a MongoDB collation field maps onto
// one of these; the mapping itself lives in the caller).
type Attribute C.UColAttribute

const (
	FrenchCollation   Attribute = C.UCOL_FRENCH_COLLATION
	AlternateHandling Attribute = C.UCOL_ALTERNATE_HANDLING
	CaseFirst         Attribute = C.UCOL_CASE_FIRST
	CaseLevel         Attribute = C.UCOL_CASE_LEVEL
	NormalizationMode Attribute = C.UCOL_NORMALIZATION_MODE
	Strength          Attribute = C.UCOL_STRENGTH
	NumericCollation  Attribute = C.UCOL_NUMERIC_COLLATION
)

// AttributeValue is the value an Attribute is set to.
type AttributeValue C.UColAttributeValue

const (
	Default AttributeValue = C.UCOL_DEFAULT
	Off     AttributeValue = C.UCOL_OFF
	On      AttributeValue = C.UCOL_ON

	// Strength.
	Primary    AttributeValue = C.UCOL_PRIMARY
	Secondary  AttributeValue = C.UCOL_SECONDARY
	Tertiary   AttributeValue = C.UCOL_TERTIARY
	Quaternary AttributeValue = C.UCOL_QUATERNARY
	Identical  AttributeValue = C.UCOL_IDENTICAL

	// CaseFirst.
	UpperFirst AttributeValue = C.UCOL_UPPER_FIRST
	LowerFirst AttributeValue = C.UCOL_LOWER_FIRST

	// AlternateHandling.
	NonIgnorable AttributeValue = C.UCOL_NON_IGNORABLE
	Shifted      AttributeValue = C.UCOL_SHIFTED
)

// ReorderCode names the group SetMaxVariable treats as the highest variable.
type ReorderCode C.UColReorderCode

const (
	ReorderSpace       ReorderCode = C.UCOL_REORDER_CODE_SPACE
	ReorderPunctuation ReorderCode = C.UCOL_REORDER_CODE_PUNCTUATION
)

// Collator is a configured ICU collator. Once opened and configured it is
// immutable, and the const compare/sort-key operations (Compare, SortKey) are
// safe for concurrent use from multiple goroutines; SetAttribute and
// SetMaxVariable are not and must be called only during setup.
type Collator struct {
	ptr *C.UCollator
}

// Open opens a collator for a bare locale (e.g. "de", "sv", "tr"; "" or "root"
// for the root/UCA order). Options are set via SetAttribute, never encoded in
// the locale string.
func Open(locale string) (*Collator, error) {
	if err := ensureData(); err != nil {
		return nil, err
	}
	cLocale := C.CString(locale)
	defer C.free(unsafe.Pointer(cLocale))
	var errBuf [64]C.char
	ptr := C.icu4c_ucol_open(cLocale, &errBuf[0], C.int(len(errBuf)))
	if ptr == nil {
		return nil, fmt.Errorf("icu4c: ucol_open(%q): %s", locale, C.GoString(&errBuf[0]))
	}
	return &Collator{ptr: ptr}, nil
}

// SetAttribute sets one collator attribute. Setup only.
func (c *Collator) SetAttribute(attr Attribute, val AttributeValue) error {
	if s := C.icu4c_ucol_set_attribute(c.ptr, C.UColAttribute(attr), C.UColAttributeValue(val)); s > 0 {
		return fmt.Errorf("icu4c: ucol_setAttribute(%d,%d): %s", attr, val, errorName(s))
	}
	return nil
}

// GetAttribute returns the collator's current value for one attribute. This
// reflects the opened locale's tailoring plus any SetAttribute overrides, so it
// exposes the resolved value (e.g. caseFirst=upper for a Danish collator that
// was never told a caseFirst) that a caller needs to validate or report.
func (c *Collator) GetAttribute(attr Attribute) (AttributeValue, error) {
	var out C.int
	if s := C.icu4c_ucol_get_attribute(c.ptr, C.UColAttribute(attr), &out); s > 0 {
		return 0, fmt.Errorf("icu4c: ucol_getAttribute(%d): %s", attr, errorName(s))
	}
	return AttributeValue(out), nil
}

// SetMaxVariable sets the highest group treated as variable under shifted
// alternate handling. Setup only.
func (c *Collator) SetMaxVariable(group ReorderCode) error {
	if s := C.icu4c_ucol_set_max_variable(c.ptr, C.UColReorderCode(group)); s > 0 {
		return fmt.Errorf("icu4c: ucol_setMaxVariable(%d): %s", group, errorName(s))
	}
	return nil
}

// Compare returns -1, 0, or 1 as a sorts before, equal to, or after b under the
// collation. It compares the Go strings as UTF-8 directly.
func (c *Collator) Compare(a, b string) int {
	var ap, bp *C.char
	if len(a) > 0 {
		ap = (*C.char)(unsafe.Pointer(unsafe.StringData(a)))
	}
	if len(b) > 0 {
		bp = (*C.char)(unsafe.Pointer(unsafe.StringData(b)))
	}
	r := int(C.icu4c_ucol_strcoll(c.ptr, ap, C.int(len(a)), bp, C.int(len(b))))
	if r == -2 {
		// A malformed-UTF-8 status is not expected from validated input; fall
		// back to a bytewise comparison so Compare never panics.
		switch {
		case a < b:
			return -1
		case a > b:
			return 1
		default:
			return 0
		}
	}
	return r
}

// Equal reports whether a and b compare equal under the collation.
func (c *Collator) Equal(a, b string) bool {
	return c.Compare(a, b) == 0
}

// SortKey returns the collation sort key for s: equal strings share a key, and
// keys order bytewise the same way the collator orders strings.
func (c *Collator) SortKey(s string) []byte {
	u16 := utf16.Encode([]rune(s))
	var src *C.UChar
	if len(u16) > 0 {
		src = (*C.UChar)(unsafe.Pointer(&u16[0]))
	}
	out := make([]byte, 256)
	n := int(C.icu4c_ucol_sortkey(c.ptr, src, C.int(len(u16)), (*C.uint8_t)(unsafe.Pointer(&out[0])), C.int(len(out))))
	if n > len(out) {
		out = make([]byte, n)
		n = int(C.icu4c_ucol_sortkey(c.ptr, src, C.int(len(u16)), (*C.uint8_t)(unsafe.Pointer(&out[0])), C.int(len(out))))
	}
	if n < 0 {
		return nil
	}
	return out[:n]
}

// Close frees the collator. A closed collator must not be used again.
func (c *Collator) Close() {
	if c.ptr != nil {
		C.ucol_close(c.ptr)
		c.ptr = nil
	}
}

func errorName(status C.int) string {
	return C.GoString(C.u_errorName(C.UErrorCode(status)))
}
