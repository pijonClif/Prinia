package download

import (
	"fmt"
	"log"
	"os"
)

// merge all downloaded sections into the final file
func MergeSections(sections []Section, finalPath string) error {

	f, err := os.OpenFile(finalPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644) //os.O_TRUNC >> truncate if exists || 0644 >> file permission
	if err != nil {
		return fmt.Errorf("open destination file: %w", err)
	}
	defer f.Close()

	for i := range sections {

		b, err := os.ReadFile(sectionPath(i))
		if err != nil {
			return fmt.Errorf("read section %d: %w", i, err)
		}
		if _, err := f.Write(b); err != nil {
			return fmt.Errorf("write section %d to destination: %w", i, err)
		}
	}

	return nil
}

// clean section up
func CleanupSections(sections []Section) {
	for i := range sections {
		if err := os.Remove(sectionPath(i)); err != nil && !os.IsNotExist(err) {
			log.Printf("warning: failed to remove temp file for section %d: %v", i, err)
		}
	}
}
