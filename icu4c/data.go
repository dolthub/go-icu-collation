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
import "C"

import (
	_ "embed"
	"fmt"
	"sync"
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

// SortKey returns the collation sort key for s under locale's default collation.
// It is a convenience wrapper around Open; callers that reuse a collation should
// Open a Collator once and call its SortKey method.
func SortKey(locale, s string) ([]byte, error) {
	c, err := Open(locale)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	return c.SortKey(s), nil
}
