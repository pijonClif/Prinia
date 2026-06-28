package download

import (
	"os"
	"path/filepath"
	"testing"
)

// writes fake section files straight to disk so we dont need
// an actual download to test the merge
func fakeSections(t *testing.T, dir string, chunks ...string) []Section {
	t.Helper()
	SetDestDir(dir)
	if err := os.MkdirAll(TempDir, 0755); err != nil {
		t.Fatalf("setup temp dir: %v", err)
	}

	sections := make([]Section, len(chunks))
	off := 0
	for i, c := range chunks {
		sections[i] = Section{Start: off, End: off + len(c) - 1}
		off += len(c)
		if err := os.WriteFile(sectionPath(i), []byte(c), 0644); err != nil {
			t.Fatalf("write fake section %d: %v", i, err)
		}
	}
	return sections
}

func TestMergeSections_Order(t *testing.T) {
	dir := t.TempDir()
	sections := fakeSections(t, dir, "hello ", "from ", "prinia")

	dest := filepath.Join(dir, "merged.txt")
	if err := MergeSections(sections, dest); err != nil {
		t.Fatalf("MergeSections: %v", err)
	}

	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read merged file: %v", err)
	}
	if string(got) != "hello from prinia" {
		t.Errorf("got %q, want %q", got, "hello from prinia")
	}
}

func TestMergeSections_MissingSection(t *testing.T) {
	dir := t.TempDir()
	// only writing section 0 to disk but telling MergeSections theres a section 1 too
	sections := fakeSections(t, dir, "only one")
	sections = append(sections, Section{Start: 8, End: 15})

	dest := filepath.Join(dir, "merged.txt")
	if err := MergeSections(sections, dest); err == nil {
		t.Error("missing section file: got nil error, want error")
	}
}

func TestMergeSections_OverwritesOldFile(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "merged.txt")

	// leftover file from some previous run, should get nuked by O_TRUNC
	if err := os.WriteFile(dest, []byte("stale data from last time"), 0644); err != nil {
		t.Fatalf("seed stale file: %v", err)
	}

	sections := fakeSections(t, dir, "fresh")
	if err := MergeSections(sections, dest); err != nil {
		t.Fatalf("MergeSections: %v", err)
	}

	got, _ := os.ReadFile(dest)
	if string(got) != "fresh" {
		t.Errorf("got %q, want %q -- old content didnt get cleared", got, "fresh")
	}
}

func TestCleanupSections_RemovesTempFiles(t *testing.T) {
	dir := t.TempDir()
	sections := fakeSections(t, dir, "a", "b", "c")

	CleanupSections(sections)

	for i := range sections {
		if _, err := os.Stat(sectionPath(i)); !os.IsNotExist(err) {
			t.Errorf("section %d still on disk after cleanup", i)
		}
	}
}

func TestCleanupSections_NothingToCleanup(t *testing.T) {
	// nothing was ever written, cleanup just shouldnt blow up
	dir := t.TempDir()
	SetDestDir(dir)
	os.MkdirAll(TempDir, 0755)

	CleanupSections([]Section{{Start: 0, End: 10}, {Start: 11, End: 20}})
}
