package icu4c

import (
	"sync"
	"testing"
)

func newColl(t *testing.T, locale string) *Collator {
	t.Helper()
	c, err := Open(locale)
	if err != nil {
		t.Fatalf("Open(%q): %v", locale, err)
	}
	t.Cleanup(c.Close)
	return c
}

func set(t *testing.T, c *Collator, a Attribute, v AttributeValue) {
	t.Helper()
	if err := c.SetAttribute(a, v); err != nil {
		t.Fatal(err)
	}
}

// Minimal-pair witnesses for known Unicode Collation Algorithm facts. The
// accented forms are explicit code points to keep the source 7-bit ASCII.
const (
	cafe      = "cafe"
	cafeAcute = "caf\u00e9" // cafe with acute e
	cafeUpper = "CAFE"
)

func TestStrengthLevels(t *testing.T) {
	// Strength 1 (primary): base letters only -- case and accent ignored.
	c := newColl(t, "en")
	set(t, c, Strength, Primary)
	if !c.Equal(cafe, cafeUpper) || !c.Equal(cafe, cafeAcute) {
		t.Errorf("primary: expected cafe == CAFE == cafe-acute")
	}

	// Strength 2 (secondary): accents distinguish, case does not.
	c = newColl(t, "en")
	set(t, c, Strength, Secondary)
	if !c.Equal(cafe, cafeUpper) {
		t.Errorf("secondary: expected cafe == CAFE (case-insensitive)")
	}
	if c.Equal(cafe, cafeAcute) {
		t.Errorf("secondary: expected cafe != cafe-acute (accent-sensitive)")
	}

	// Strength 3 (tertiary): case distinguishes too.
	c = newColl(t, "en")
	set(t, c, Strength, Tertiary)
	if c.Equal(cafe, cafeUpper) {
		t.Errorf("tertiary: expected cafe != CAFE (case-sensitive)")
	}
	if c.Equal(cafe, cafeAcute) {
		t.Errorf("tertiary: expected cafe != cafe-acute")
	}
}

func TestCaseLevel(t *testing.T) {
	// At strength 2, cafe == CAFE. Adding a case level re-separates them while
	// still ignoring accents at the primary/secondary levels.
	c := newColl(t, "en")
	set(t, c, Strength, Secondary)
	set(t, c, CaseLevel, On)
	if c.Equal(cafe, cafeUpper) {
		t.Errorf("secondary+caseLevel: expected cafe != CAFE")
	}
}

func TestCaseFirst(t *testing.T) {
	c := newColl(t, "en")
	set(t, c, Strength, Tertiary)
	set(t, c, CaseFirst, UpperFirst)
	if c.Compare("A", "a") >= 0 {
		t.Errorf("caseFirst=upper: expected A < a")
	}

	c = newColl(t, "en")
	set(t, c, Strength, Tertiary)
	set(t, c, CaseFirst, LowerFirst)
	if c.Compare("a", "A") >= 0 {
		t.Errorf("caseFirst=lower: expected a < A")
	}
}

func TestNumericOrdering(t *testing.T) {
	// Without numeric ordering, "a10" < "a2" lexically ('1' < '2').
	c := newColl(t, "en")
	set(t, c, Strength, Tertiary)
	if c.Compare("a10", "a2") >= 0 {
		t.Errorf("no numeric: expected a10 < a2 lexically")
	}
	// With numeric ordering, "a2" < "a10" numerically.
	c = newColl(t, "en")
	set(t, c, Strength, Tertiary)
	set(t, c, NumericCollation, On)
	if c.Compare("a2", "a10") >= 0 {
		t.Errorf("numeric: expected a2 < a10")
	}
}

func TestAlternateShifted(t *testing.T) {
	blackBirdHyphen, blackbird := "black-bird", "blackbird"

	// Default (non-ignorable): the hyphen is significant, so they differ.
	c := newColl(t, "en")
	set(t, c, Strength, Tertiary)
	if c.Equal(blackBirdHyphen, blackbird) {
		t.Errorf("non-ignorable: expected black-bird != blackbird")
	}

	// Shifted at strength 3: punctuation becomes ignorable, so they are equal.
	c = newColl(t, "en")
	set(t, c, Strength, Tertiary)
	set(t, c, AlternateHandling, Shifted)
	if !c.Equal(blackBirdHyphen, blackbird) {
		t.Errorf("shifted/tertiary: expected black-bird == blackbird")
	}

	// Shifted at strength 4 (quaternary): the ignored punctuation is compared at
	// the quaternary level, separating them again.
	c = newColl(t, "en")
	set(t, c, Strength, Quaternary)
	set(t, c, AlternateHandling, Shifted)
	if c.Equal(blackBirdHyphen, blackbird) {
		t.Errorf("shifted/quaternary: expected black-bird != blackbird")
	}
}

func TestFrenchBackwards(t *testing.T) {
	// The classic French quartet differs only in accents. French collation
	// compares the secondary (accent) level right-to-left, which flips the order
	// of cote-acute and cote-circumflex relative to the forward comparison.
	coteAcute := "cot\u00e9"    // cote with acute e
	coteCircum := "c\u00f4te"   // cote with circumflex o
	forward := newColl(t, "fr") // fr default is forward secondary
	set(t, forward, Strength, Tertiary)
	set(t, forward, FrenchCollation, Off)

	backward := newColl(t, "fr")
	set(t, backward, Strength, Tertiary)
	set(t, backward, FrenchCollation, On)

	f := forward.Compare(coteAcute, coteCircum)
	b := backward.Compare(coteAcute, coteCircum)
	if f == 0 || b == 0 {
		t.Fatalf("expected a strict order both ways (f=%d b=%d)", f, b)
	}
	if (f < 0) == (b < 0) {
		t.Errorf("backwards: expected the order of cote-acute vs cote-circumflex to flip (forward=%d backward=%d)", f, b)
	}
}

func TestLocaleTailoringDiffersFromRoot(t *testing.T) {
	// Swedish sorts a-ring, a-umlaut, o-umlaut after z; the root order does not.
	aRing, z := "\u00e5", "z"
	root := newColl(t, "")
	set(t, root, Strength, Primary)
	sv := newColl(t, "sv")
	set(t, sv, Strength, Primary)
	if !(root.Compare(aRing, z) < 0) {
		t.Errorf("root: expected a-ring < z")
	}
	if !(sv.Compare(aRing, z) > 0) {
		t.Errorf("sv: expected a-ring > z (Swedish tailoring)")
	}
}

func TestUnknownLocaleAccepted(t *testing.T) {
	// ICU does not error on an unknown locale; it falls back to root. (MongoDB's
	// stricter rejection is enforced by the caller, not the binding.)
	c, err := Open("zz-nonsense")
	if err != nil {
		t.Fatalf("Open(unknown): unexpected error %v", err)
	}
	c.Close()
}

func TestSortKeyOrdersLikeCompare(t *testing.T) {
	c := newColl(t, "en")
	set(t, c, Strength, Tertiary)
	words := []string{"apple", "Apple", cafe, cafeAcute, cafeUpper, "banana", "a2", "a10"}
	for _, a := range words {
		for _, b := range words {
			cmp := c.Compare(a, b)
			ka, kb := c.SortKey(a), c.SortKey(b)
			key := bytesCompare(ka, kb)
			if (cmp < 0) != (key < 0) || (cmp == 0) != (key == 0) || (cmp > 0) != (key > 0) {
				t.Errorf("SortKey order disagrees with Compare for %q vs %q: cmp=%d key=%d", a, b, cmp, key)
			}
		}
	}
}

func bytesCompare(a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			if a[i] < b[i] {
				return -1
			}
			return 1
		}
	}
	switch {
	case len(a) < len(b):
		return -1
	case len(a) > len(b):
		return 1
	default:
		return 0
	}
}

// TestConcurrentUse hammers a single configured collator from many goroutines
// through the const operations (Compare, SortKey), confirming the shared-collator
// design under -race.
func TestConcurrentUse(t *testing.T) {
	c := newColl(t, "en")
	set(t, c, Strength, Tertiary)
	words := []string{"cafe", cafeUpper, cafeAcute, "apple", "Apple", "banana", "a2", "a10"}

	var wg sync.WaitGroup
	for g := 0; g < 16; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 2000; i++ {
				a := words[i%len(words)]
				b := words[(i+7)%len(words)]
				_ = c.Compare(a, b)
				_ = c.SortKey(a)
			}
		}()
	}
	wg.Wait()
}
