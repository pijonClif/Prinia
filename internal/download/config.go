package download

import "path/filepath"

var (
	DestDir string
	TempDir string
)

// configures the download directory defaults to "./downloads"
func SetDestDir(dir string) {
	if dir == "" {
		dir = "downloads"
	}
	DestDir = dir
	TempDir = filepath.Join(DestDir, "sections")
}
