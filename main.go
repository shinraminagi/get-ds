package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var intervalFlag = flag.Float64("interval", 1, "Interval between each download (sec)")

func main() {
	flag.Parse()
	url := flag.Arg(0)

	if !regexp.MustCompile(`^http://ddd-smart\.net/dl-m-m\.php\?cn=\d+/\d+`).MatchString(url) {
		log.Fatalln("Invalid Dojin Smart URL")
	}

	fmt.Printf("Scraping %s...", url)
	res, err := http.Get(url)
	if err != nil {
		log.Fatalf("Couldn't retreive %s: %s\n", url, err)
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		log.Fatalf("Couldn't parse response body: %s\n", err)
	}
	fmt.Println("done")

	doc.Find(`div#comic-area > img`).Each(func(_ int, el *goquery.Selection) {
		for {
			imgUrl, ok := el.Attr("src")
			if !ok {
				break
			}

			if *intervalFlag > 0 {
				fmt.Printf("Waiting for %f seconds...", *intervalFlag)
				time.Sleep(time.Duration(*intervalFlag) * time.Second)
				fmt.Println("OK.")
			}

			fmt.Printf("Downloading %s ...", imgUrl)
			err = download(imgUrl)
			if err != nil {
				log.Println()
				log.Println(err)
				fmt.Println("Retry...")
				continue
			}
			fmt.Println("OK.")
			break
		}
	})
}

func download(rawurl string) error {
	filename, err := fileNameOf(rawurl)
	if err != nil {
		return err
	}
	resp, err := http.Get(rawurl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

var reInPath = regexp.MustCompile("[^/]+$")

func fileNameOf(rawurl string) (string, error) {
	url, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	file := reInPath.FindString(url.Path)
	if file == "" {
		return "", fmt.Errorf("Filename not found: %s", rawurl)
	}
	return file, nil
}
