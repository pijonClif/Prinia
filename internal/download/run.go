package download

import (
	"fmt"
	"os"
	"time"
)

const (
	TempDir = "downloads/sections/" //directory for temp files
	DestDir = "downloads/"          //dir for final file
)

func EnsureDirs() error {
	for _, d := range []string{DestDir, TempDir} {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("create dir %s: %w", d, err)
		}
	}
	return nil
}

func Run(url, fileName string, numSections int) error {
	start := time.Now()

	if err := EnsureDirs(); err != nil {
		return err
	}

	size, err := FileSize(url)
	if err != nil {
		return fmt.Errorf("get file size: %w", err)
	}

	sections := CalcSections(size, numSections)

	if err := DownloadAll(sections, url); err != nil {
		CleanupSections(sections)
		return fmt.Errorf("download sections: %w", err)
	}

	destPath := DestPath(fileName)
	if err := MergeSections(sections, destPath); err != nil {
		CleanupSections(sections)
		return fmt.Errorf("merge sections: %w", err)
	}

	CleanupSections(sections)

	fmt.Println("download URL:", url)
	fmt.Println("saved to:", destPath)
	fmt.Println("elapsed:", time.Since(start))
	return nil
}
