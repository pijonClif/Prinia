package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/h2non/filetype"
)

const (
	TempDir = "downloads/sections/" //directory for temp files
	DestDir = "downloads/"          //dir for final file
)

// tried to MIME the filetype from the header
// evil
// inacc
// we now go with h2non/filetype
// glory to h2non
func sniffExt(url string) string {
	req, err := NewRequest(http.MethodGet, url)
	if err != nil {
		return ""
	}
	req.Header.Set("Range", "bytes=0-260")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	buf := make([]byte, 261)            //reads the bytes [0, 260] -- magic bytes ((we learn something new)) || "magic bytes are the first bytes of a file that help identify its type"
	n, _ := io.ReadFull(resp.Body, buf) //match agaiinst known signatures

	kind, err := filetype.Match(buf[:n])
	if err != nil || kind == filetype.Unknown {
		return ""
	}
	return "." + kind.Extension //append
}

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

	if path.Ext(fileName) == "" {
		if ext := sniffExt(url); ext != "" {
			fileName += ext
		}
	}

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
