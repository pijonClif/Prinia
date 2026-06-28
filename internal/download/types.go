package download

import (
	"fmt"
	"path/filepath"
)

const (
	userAgent = "Prinia"
)

// section >> half open byte range [Start, End] ((inclusive, like HTTP Range))
type Section struct {
	Start, End int
}

func sectionPath(i int) string {
	return filepath.Join(TempDir, fmt.Sprintf("section-%d.tmp", i))
}

// DestPath joins the dest dir with a filename.
func DestPath(fileName string) string {
	return filepath.Join(DestDir, fileName)
}
