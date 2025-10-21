package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

/*======== TO-DO ========
>> [o] user input ((flag))
>> [-] download multiple sources
>> [-] progress display
>> [-] a frontend maybe?? ((fyne))
*/

const (
	TempDir   = "downloads/sections/" //directory for temp files
	DestDir   = "downloads/"          //dir for final file
	userAgent = "Dinky Doinks"

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

// getes user  innput; returns fileURL and dest file name
func userInput() (DownloadTask, error) {

	var (
		d           DownloadTask
		inputURL    string
		fileName    string
		ttlSections int
	)

	flag.StringVar(&inputURL, "u", "", "download URL")
	flag.StringVar(&fileName, "f", "", "file name")
	flag.IntVar(&ttlSections, "s", 0, "num of sections")

	flag.Parse()

	if inputURL == "" || fileName == "" || ttlSections == 0 {
		flag.Usage()
		return DownloadTask{}, fmt.Errorf("missing arguments")
	}

	d.URL = inputURL
	d.FileName = fileName
	d.Sections = ttlSections

	return d, nil
}

// returns new http req
func getNewRequest(method, url string) (*http.Request, error) {
	req, err := http.NewRequest(
		method,
		url,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("\nfailed to create new request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	return req, nil
}

// returns total file size
func getFileSize(url string) (int, error) {

	req, err := getNewRequest("HEAD", url)
	if err != nil {
		return 0, fmt.Errorf("\nfailed to create HEAD request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("\nfailed to execute HEAD request: %w", err)
	}
	defer resp.Body.Close()

	fmt.Println("Status Code: ", resp.StatusCode)

	if resp.StatusCode > 299 {
		return 0, nil
	}

	fileSize, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return 0, fmt.Errorf("\nfailed to convert Content-Length to integer: %w", err)
	}

	return fileSize, nil
}

// diviides ttl file into sections; returns section sizes [byteStart byteEnd]
func calcSections(fileSize, ttlSections int) [][2]int {

	//sections
	var sections = make([][2]int, ttlSections)
	eachSection := fileSize / ttlSections

	fmt.Printf("each size is %v bytes", eachSection)

	for i := range sections {

		if i == 0 {
			//starting byte of first section
			sections[i][0] = 0
		} else {
			//starting byte of other sections
			sections[i][0] = sections[i-1][1] + 1
		}

		if i < ttlSections-1 {
			//ending byte of other sections
			sections[i][1] = sections[i][0] + eachSection
		} else {
			//ending byte of last section
			sections[i][1] = fileSize - 1
		}
	}
	fmt.Println(sections)
	return sections
}

// download a section of the file; return bytes downloaded
func dwnldSection(i int, s [2]int, url string) (int, error) {

	req, err := getNewRequest("GET", url)
	if err != nil {
		return 0, fmt.Errorf("\nfailed to create GET request: %w", err)
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", s[0], s[1]))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("\nfailed to execute GET request: %w", err)
	}
	defer resp.Body.Close()

	//fmt.Printf("downloaded %v bytes for section %v: %v\n", resp.Header.Get("Content-Length"), i, s)

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("\nfailed to read response body: %w", err)
	}

	err = os.WriteFile(TempDir+fmt.Sprintf("section-%v.tmp", i), b, os.ModePerm)
	if err != nil {
		return 0, fmt.Errorf("\nfailed to write section file: %w", err)
	}

	return len(b), nil
}

// concurrently downloads all sections
func dwnldSectionsConc(sections [][2]int, url string) error {

	var wg sync.WaitGroup
	for i, s := range sections {
		wg.Add(1)
		//values fixed:
		i := i
		s := s
		go func() {
			defer wg.Done()
			bytesDwnld, err := dwnldSection(i, s, url)
			if err != nil {
				log.Println("err")
				panic(err)
			}
			fmt.Printf("downloaded %v bytes for section %v: %v\n", bytesDwnld, i, s)
		}()
	}
	wg.Wait()
	fmt.Printf("All sectiond downloaded")

	return nil
}

// merge all downloaded sections into the final file
func dwnldMerge(sections [][2]int, fnlFilePath string) error {

	f, err := os.OpenFile(fnlFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		return fmt.Errorf("\nfailed to open target file: %w", err)
	}
	defer f.Close()

	for i := range sections {

		b, err := os.ReadFile(TempDir + fmt.Sprintf("section-%v.tmp", i))
		if err != nil {
			return fmt.Errorf("\nfailed to read section file: %w", err)
		}

		n, err := f.Write(b)
		if err != nil {
			return fmt.Errorf("\nfailed to write to target file: %w", err)
		}
		fmt.Printf("\n%v bytes merged", n)
	}

	return nil
}

// delete section ((after merge))
func delSections(sections [][2]int, TargetPath string) error {

	if _, err := os.Stat(TempDir); err == nil {

		for i := range sections {
			err := os.Remove(TempDir + fmt.Sprintf("section-%v.tmp", i))
			if err != nil {
				return fmt.Errorf("\nFaiiled to delete target file: %w", err)
			}
		}
	} /* else if os.IsNotExist(err) {
		fmt.Sprintf("\nFile %v does not exist: %w", DestDir+TargetPath, err)
	} else {
		fmt.Sprintf("\nERR: Schrodinger's file")
	}
	*/

	return nil
}

func main() {
	starTime := time.Now()

	//---
	if _, err := os.Stat(DestDir); os.IsNotExist(err) {
		err = os.MkdirAll(DestDir, os.ModeDir|0755)
		if err != nil {
			log.Fatalf("Failed to create destination directory: %v", err)
		}
	}
	if _, err := os.Stat(TempDir); os.IsNotExist(err) {
		err = os.MkdirAll(TempDir, os.ModeDir|0755)
		if err != nil {
			log.Fatalf("Failed to create temporary directory: %v", err)
		}
	}
	//---

	d, err := userInput()
	if err != nil {
		log.Fatalf("Failed to retrieve user input: %v", err)
	}

	fileSize, err := getFileSize(d.URL)
	if err != nil {
		log.Fatalf("Failed to get file size: %v", err)
	}

	sections := calcSections(fileSize, d.Sections)

	err = dwnldSectionsConc(sections, d.URL)
	if err != nil {
		log.Fatalf("Failed to download sections: %v", err)
	}

	fnlFilePath := DestDir + d.FileName

	err = dwnldMerge(sections, fnlFilePath)
	if err != nil {
		log.Fatalf("Failed to merge sections: %v", err)
	}

	err = delSections(sections, fnlFilePath)
	if err != nil {
		log.Fatalf("Failed to delete sections: %v", err)
	}

	fmt.Println("\ndownload URL: ", d.URL)
	fmt.Println("download destination: ", fnlFilePath)
	fmt.Println("return time: ", time.Since(starTime))
}
