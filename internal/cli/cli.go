package cli

import (
	"flag"
	"fmt"
)

type DownloadTask struct {
	URL      string
	FileName string
	Sections int
	Dir      string
	DirOnly  bool // true: user passed -dir alone // just set up the folder skip download
}

// read n validate cli args
func ParseFlags() (DownloadTask, error) {

	var (
		inputURL    string
		fileName    string
		ttlSections int
		dir         string
	)

	flag.StringVar(&inputURL, "u", "", "download URL")
	flag.StringVar(&fileName, "f", "", "output file name")
	flag.IntVar(&ttlSections, "s", 0, "number of sections")
	flag.StringVar(&dir, "dir", "", "download directory (default: ./downloads)")

	flag.Parse()

	task := DownloadTask{URL: inputURL, FileName: fileName, Sections: ttlSections, Dir: dir}

	// -dir given on its own: just configure the directory, nothing else required
	if dir != "" && inputURL == "" && fileName == "" && ttlSections <= 0 {
		task.DirOnly = true
		return task, nil
	}

	if inputURL == "" || fileName == "" || ttlSections <= 0 {
		flag.Usage()
		return DownloadTask{}, fmt.Errorf("missing arguments")
	}

	return task, nil
}
