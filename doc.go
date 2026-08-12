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

// Package icu is a cgo binding to ICU's ucol_* collation API. It vendors a
// pinned ICU built statically, so collation behavior is fixed to that ICU
// version rather than whatever the host ships. It binds only the collation
// surface (open a collator, set attributes, compare strings, produce sort
// keys); it is independent of go-icu-regex, which links the system ICU and
// binds only the regex API.
//
// The binding itself is added in a later change; this file establishes the
// module and its root package.
package icu
