package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

/*======== TO-DO ========
>> [o] user input ((flag))
>> [~] bug fix + upar-upar se redo err handling
>> [~] refactrr code
>> [-] download multiple sources
>> [-] progress display
>> [-] a frontend maybe?? ((fyne))
*/

const (
	TempDir   = "downloads/sections/" //directory for temp files
	DestDir   = "downloads/"          //dir for final file
	userAgent = "Prinia"

	/*
		Url1     = "https://starecat.com/content/wp-content/uploads/tired-cat-smoking-a-cigarette.jpg"
		Url2     = "https://media1.tenor.com/m/oGXvGs3Lp08AAAAC/meme-cat.gif"
		Url500MB = "https://mmatechnical.com/Download/Download-Test-File/(MMA)-500MB.zip"
		Url1GB   = "https://mmatechnical.com/Download/Download-Test-File/(MMA)-1GB.zip"
	*/
)

type DownloadTask struct {
	URL      string
	FileName string
	Sections int
}

// section >> half open byte range [Start, End] ((inclusive, like HTTP Range))
type section struct {
	Start, End int
}

// read n validate cli args
func parseFlags() (DownloadTask, error) {

	var (
		inputURL    string
		fileName    string
		ttlSections int
	)

	flag.StringVar(&inputURL, "u", "", "download URL")
	flag.StringVar(&fileName, "f", "", "output file name")
	flag.IntVar(&ttlSections, "s", 0, "number of sections")

	flag.Parse()

	if inputURL == "" || fileName == "" || ttlSections <= 0 {
		flag.Usage()
		return DownloadTask{}, fmt.Errorf("missing arguments")
	}

	return DownloadTask{URL: inputURL, FileName: fileName, Sections: ttlSections}, nil
}

// returns new http req
func newRequest(method, url string) (*http.Request, error) {
	req, err := http.NewRequest(
		method,
		url,
		nil,
	)

	if err != nil {
		return nil, fmt.Errorf("\ncreate %s request: %w", method, err)
	}

	req.Header.Set("User-Agent", userAgent)
	return req, nil
}

// head req >> fetch ocntent length
func fileSize(url string) (int, error) {

	req, err := newRequest(http.MethodHead, url)
	if err != nil {
		return 0, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("HEAD request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return 0, fmt.Errorf("HEAD request returned status %d", resp.StatusCode)
	}

	cl := resp.Header.Get("Content-Length")
	if cl == "" {
		return 0, fmt.Errorf("server did not return Content-Length (chunked or unknown size)")
	}

	size, err := strconv.Atoi(cl)
	if err != nil {
		return 0, fmt.Errorf("parse Content-Length: %w", err)
	}

	return size, nil
}

// diviides ttl file into sections; returns section sizes [byteStart byteEnd]
func calcSections(size, n int) []section {

	sections := make([]section, n)
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

func sectionPath(i int) string {
	return filepath.Join(TempDir, fmt.Sprintf("section-%d.tmp", i))
}

// download a section of the file; return bytes downloaded
func downloadSection(i int, s section, url string) (int, error) {

	req, err := newRequest(http.MethodGet, url)
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
func downloadAll(sections []section, url string) error {
	var (
		wg       sync.WaitGroup
		mu       sync.Mutex // protexts shared data (firstErr)
		firstErr error      // first err encounterd
	)

	for i, s := range sections {
		wg.Add(1)

		go func(i int, s section) {
			defer wg.Done()
			n, err := downloadSection(i, s, url)
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

// merge all downloaded sections into the final file
func mergeSections(sections []section, finalPath string) error {

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
func cleanupSections(sections []section) {
	for i := range sections {
		if err := os.Remove(sectionPath(i)); err != nil && !os.IsNotExist(err) {
			log.Printf("warning: failed to remove temp file for section %d: %v", i, err)
		}
	}
}

// the functubn name bruh
func ensureDirs() error {
	for _, d := range []string{DestDir, TempDir} {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("create dir %s: %w", d, err)
		}
	}
	return nil
}

func run() error {
	start := time.Now()

	if err := ensureDirs(); err != nil {
		return err
	}

	task, err := parseFlags()
	if err != nil {
		return err
	}

	size, err := fileSize(task.URL)
	if err != nil {
		return fmt.Errorf("get file size: %w", err)
	}

	sections := calcSections(size, task.Sections)

	if err := downloadAll(sections, task.URL); err != nil {
		cleanupSections(sections)
		return fmt.Errorf("download sections: %w", err)
	}

	destPath := filepath.Join(DestDir, task.FileName)
	if err := mergeSections(sections, destPath); err != nil {
		cleanupSections(sections)
		return fmt.Errorf("merge sections: %w", err)
	}

	cleanupSections(sections)

	fmt.Println("download URL:", task.URL)
	fmt.Println("saved to:", destPath)
	fmt.Println("elapsed:", time.Since(start))
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
