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

// Package icu4c is the low-level cgo binding to the vendored, statically-built
// ICU. It compiles ICU's common and i18n C++ sources directly (see the flat
// .cpp/.h files in this directory), so the collation engine is pinned to this
// ICU version regardless of any system ICU.
package icu4c

// #cgo CPPFLAGS: -I${SRCDIR} -DU_STATIC_IMPLEMENTATION=1 -DU_COMMON_IMPLEMENTATION=1 -DU_I18N_IMPLEMENTATION=1 -DU_CHARSET_IS_UTF8=1
// #cgo CXXFLAGS: -std=c++17
// #cgo LDFLAGS: -lstdc++ -lm
// #include "unicode/uversion.h"
//
// static void icu4c_version(char *out) {
//     UVersionInfo v;
//     u_getVersion(v);
//     u_versionToString(v, out);
// }
import "C"

// Version returns the linked ICU version string, e.g. "78.3".
func Version() string {
	var buf [C.U_MAX_VERSION_STRING_LENGTH]C.char
	C.icu4c_version(&buf[0])
	return C.GoString(&buf[0])
}
