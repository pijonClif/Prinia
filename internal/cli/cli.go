package cli

import (
	"flag"
	"fmt"
)

type DownloadTask struct {
	URL      string
	FileName string
	Sections int
}

// read n validate cli args
func ParseFlags() (DownloadTask, error) {

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
