package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

// diviides ttl file into sections; returns section sizes [byteStart byteEnd]
func CalcSections(size, n int) []Section {

	sections := make([]Section, n)
	each := size / n

	for i := range sections {

		if i == 0 {
			//starting byte of first section
			sections[i].Start = 0
		} else {
			//starting byte of other sections
			sections[i].Start = sections[i-1].End + 1
		}

		if i < n-1 {
			//ending byte of other sections
			sections[i].End = sections[i].Start + each - 1
		} else {
			//ending byte of last section
			sections[i].End = size - 1
		}
	}
	return sections
}

func DownloadSection(i int, s Section, url string) (int, error) {

	req, err := NewRequest(http.MethodGet, url)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", s.Start, s.End))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("section %d: GET request: %w", i, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("section %d: unexpected status %d", i, resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("section %d: read body: %w", i, err)
	}

	if err := os.WriteFile(sectionPath(i), b, 0644); err != nil {
		return 0, fmt.Errorf("section %d: write temp file: %w", i, err)
	}

	return len(b), nil
}

// concurrently downloads all sections
func DownloadAll(sections []Section, url string) error {
	var (
		wg       sync.WaitGroup
		mu       sync.Mutex // protexts shared data (firstErr)
		firstErr error      // first err encounterd
	)

	for i, s := range sections {
		wg.Add(1)

		go func(i int, s Section) {
			defer wg.Done()
			n, err := DownloadSection(i, s, url)
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
				return
			}
			fmt.Printf("section %d: %d bytes (%d-%d)\n", i, n, s.Start, s.End)

		}(i, s)
	}
	wg.Wait()

	return firstErr
}
