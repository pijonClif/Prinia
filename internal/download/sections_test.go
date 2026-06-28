package download

import "testing"

func TestCalcSections_Even(t *testing.T) {
	// 1000 bytes, 4 sections -- should divide clean, no remainder
	s := CalcSections(1000, 4)

	if len(s) != 4 {
		t.Fatalf("got %d sections, want 4", len(s))
	}

	want := []Section{
		{Start: 0, End: 249},
		{Start: 250, End: 499},
		{Start: 500, End: 749},
		{Start: 750, End: 999},
	}
	for i := range want {
		if s[i] != want[i] {
			t.Errorf("section %d = %+v, want %+v", i, s[i], want[i])
		}
	}
}

func TestCalcSections_Uneven(t *testing.T) {
	// 1000 / 3 doesnt divide clean, last section eats the remainder
	s := CalcSections(1000, 3)

	if len(s) != 3 {
		t.Fatalf("got %d sections, want 3", len(s))
	}

	// every section should start right where the last one ended,
	// no gaps no overlap
	for i := 1; i < len(s); i++ {
		if s[i].Start != s[i-1].End+1 {
			t.Errorf("section %d starts at %d, section %d ended at %d -- gap or overlap", i, s[i].Start, i-1, s[i-1].End)
		}
	}

	// last section has to reach the actual end of the file
	if s[len(s)-1].End != 999 {
		t.Errorf("last section ends at %d, want 999", s[len(s)-1].End)
	}
}

func TestCalcSections_OneSection(t *testing.T) {
	// n=1 -> whole file is just one section, basically no-op
	s := CalcSections(500, 1)

	if len(s) != 1 || s[0].Start != 0 || s[0].End != 499 {
		t.Errorf("got %+v, want a single section {0 499}", s)
	}
}

// known issue: if you ask for more sections than there are bytes,
// CalcSections doesn't guard against it and you get garbage sections
// near the end (Start > End). not fixing it here, just documenting
// what currently happens so nobody's surprised later.
func TestCalcSections_MoreSectionsThanBytes(t *testing.T) {
	s := CalcSections(5, 10)

	if len(s) != 10 {
		t.Fatalf("got %d sections, want 10", len(s))
	}
	if s[len(s)-1].End != 4 {
		t.Errorf("last section should end at byte 4, got %d", s[len(s)-1].End)
	}
}
