package main

import (
	"log"

	"prinia/internal/cli"
	"prinia/internal/download"
)

/*======== TO-DO ========
>> [o] user input ((flag))
>> [o] bug fix + upar-upar se redo err handling
>> [o] refactrr code
>> [-] download multiple sources
>> [-] progress display
>> [-] a frontend maybe?? ((fyne))
>> [-] tests apparently*/

/*
const (
	Url1     = "https://starecat.com/content/wp-content/uploads/tired-cat-smoking-a-cigarette.jpg"
	Url2     = "https://media1.tenor.com/m/oGXvGs3Lp08AAAAC/meme-cat.gif"
	Url500MB = "https://mmatechnical.com/Download/Download-Test-File/(MMA)-500MB.zip"
	Url1GB   = "https://mmatechnical.com/Download/Download-Test-File/(MMA)-1GB.zip"
)
*/

func main() {
	task, err := cli.ParseFlags()
	if err != nil {
		log.Fatal(err)
	}

	download.SetDestDir(task.Dir)

	if task.DirOnly {
		if err := download.EnsureDirs(); err != nil {
			log.Fatal(err)
		}
		log.Printf("download directory ready: %s", download.DestDir)
		return
	}

	if err := download.Run(task.URL, task.FileName, task.Sections); err != nil {
		log.Fatal(err)
	}
}
