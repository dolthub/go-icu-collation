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

// #include "unicode/udata.h"
// #include "unicode/ucol.h"
// #include "unicode/utypes.h"
// #include <stdlib.h>
// #include <string.h>
//
// // Register the embedded ICU data blob as the process common data. ICU keeps
// // the pointer without copying, so the buffer must outlive the process; a
// // malloc'd copy (16-byte aligned on glibc, as ICU requires) is never freed.
// static const char *icu4c_set_common_data(const void *data, int len) {
//     UErrorCode status = U_ZERO_ERROR;
//     void *buf = malloc((size_t)len);
//     if (buf == NULL) {
//         return "malloc failed";
//     }
//     memcpy(buf, data, (size_t)len);
//     udata_setCommonData(buf, &status);
//     if (U_FAILURE(status)) {
//         return u_errorName(status);
//     }
//     udata_setFileAccess(UDATA_NO_FILES, &status);
//     return "";
// }
//
// // Compute the ICU sort key for a UTF-16 string under locale's collation into
// // out (capacity cap). Returns the key length, or -1 if the collator fails to
// // open (with u_errorName written to err). Warnings (root/parent fallback) are
// // not failures.
// static int icu4c_sortkey(const char *locale, const UChar *s, int slen,
//                          uint8_t *out, int cap, char *err, int errcap) {
//     UErrorCode status = U_ZERO_ERROR;
//     UCollator *coll = ucol_open(locale, &status);
//     if (coll == NULL || U_FAILURE(status)) {
//         if (err != NULL && errcap > 0) {
//             const char *name = u_errorName(status);
//             int i = 0;
//             for (; name[i] != 0 && i < errcap - 1; i++) err[i] = name[i];
//             err[i] = 0;
//         }
//         if (coll != NULL) {
//             ucol_close(coll);
//         }
//         return -1;
//     }
//     int n = ucol_getSortKey(coll, s, slen, out, cap);
//     ucol_close(coll);
//     return n;
// }
import "C"

import (
	_ "embed"
	"fmt"
	"sync"
	"unicode/utf16"
	"unsafe"
)

//go:embed icudt78l.dat
var icuData []byte

var (
	dataOnce sync.Once
	dataErr  error
)

// ensureData registers the embedded ICU data with ICU on first use. It is safe
// to call repeatedly and from multiple goroutines.
func ensureData() error {
	dataOnce.Do(func() {
		if len(icuData) == 0 {
			dataErr = fmt.Errorf("icu4c: no ICU data embedded")
			return
		}
		msg := C.GoString(C.icu4c_set_common_data(unsafe.Pointer(&icuData[0]), C.int(len(icuData))))
		if msg != "" {
			dataErr = fmt.Errorf("icu4c: udata_setCommonData: %s", msg)
		}
	})
	return dataErr
}

// SortKey returns the ICU collation sort key for s under locale's collation.
// Two strings equal under the collation share a key, and keys order bytewise the
// same way the collator orders strings. It is the primitive used to fingerprint
// a locale's collation, so a trimmed data blob can be checked against the full
// one for every locale.
func SortKey(locale, s string) ([]byte, error) {
	if err := ensureData(); err != nil {
		return nil, err
	}
	cLocale := C.CString(locale)
	defer C.free(unsafe.Pointer(cLocale))

	u16 := utf16.Encode([]rune(s))
	var src *C.UChar
	if len(u16) > 0 {
		src = (*C.UChar)(unsafe.Pointer(&u16[0]))
	}

	var errBuf [64]C.char
	out := make([]byte, 256)
	n := C.icu4c_sortkey(cLocale, src, C.int(len(u16)), (*C.uint8_t)(unsafe.Pointer(&out[0])), C.int(len(out)), &errBuf[0], C.int(len(errBuf)))
	if n < 0 {
		return nil, fmt.Errorf("icu4c: ucol_open(%q) failed: %s", locale, C.GoString(&errBuf[0]))
	}
	if int(n) > len(out) {
		out = make([]byte, int(n))
		n = C.icu4c_sortkey(cLocale, src, C.int(len(u16)), (*C.uint8_t)(unsafe.Pointer(&out[0])), C.int(len(out)), &errBuf[0], C.int(len(errBuf)))
		if n < 0 {
			return nil, fmt.Errorf("icu4c: ucol_open(%q) failed: %s", locale, C.GoString(&errBuf[0]))
		}
	}
	return out[:n], nil
}
