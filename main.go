package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var downloadDir = "/tmp/image_down"
var numWorkers = 100
var bufferedChannelSize = 50000
var myWaitGroup sync.WaitGroup

func init() {
	log.SetOutput(os.Stdout) //stdout not stderr
}

//main drives the process -
// takes one argument for the input file with URLs to download
// uses a channel to feed go routine which does the download concurrently
func main() {
	log.Println("process begins")

	var inputFile string
	if len(os.Args) > 1 {
		inputFile = os.Args[1]
		log.Println("input file name: ", inputFile)
	} else {
		log.Fatal("argument is missing - which input file has URLs?")
	}

	urls, err := getUrls(inputFile)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("count of urls to download: ", len(urls))
	log.Println("number of workers: ", numWorkers)
	log.Println("size of buffered channel: ", bufferedChannelSize)
	log.Println("download destination: ", downloadDir)

	err = os.MkdirAll(downloadDir, os.ModePerm) //create download dir if not there
	if err != nil {
		log.Fatal(err)
	}

	myWaitGroup.Add(numWorkers) //tell waitgroup about the Number of Workers
	channelU := make(chan string, bufferedChannelSize)

	for w := 1; w <= numWorkers; w++ {
		go downloadFile(channelU) //start go routines up to total number of workers
	}

	start := time.Now()

	log.Println("writing URLs to the channel")

	for _, url := range urls {
		channelU <- url //put every URL string into the channel
	}
	close(channelU) //close the channel since nothing more will be written to it

	log.Println("closed the channel")

	myWaitGroup.Wait() //block until waitgroup is finished

	elapsed := time.Since(start).Minutes()
	log.Println("elapsed minutes", elapsed)

	log.Println("process is complete")
}

// extract rightmost part of a URL, and prepend with a directory name. Useful
// to take a URL (like an image), and choose where to download to.
//  e.g. input 'https://yahoo.com/marissa.jpg' yields '/tmp/foo/marissa.jpg'
func fileFromURL(url string) string {
	lastSlash := strings.LastIndex(url, "/")
	target := downloadDir + url[lastSlash+1:] //same as url[lastSlash+1 : len(url)]
	return target
}

// reads a file into memory and returns a slice of its lines.
func getUrls(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// read URL from a channel, and download from URL to disk
func downloadFile(channelU chan string) {
	defer myWaitGroup.Done()

	//forever loop - until channel read comes up empty
	for {
		url, ok := <-channelU
		if !ok {
			return // channel is empty and closed, worker to end
		}

		saveToPath := fileFromURL(url)
		out, err := os.Create(saveToPath)
		if err != nil {
			log.Fatal(err)
		}

		res, httpError := http.Get(url)
		if httpError == nil {
			_, err = io.Copy(out, res.Body) // Write the body to file
			if err != nil {
				fmt.Println(err) // or log.Fatal(err)
			} 
			out.Close()
			res.Body.Close()
		} else {
			log.Println(httpError) // or else log.Fatal(err) or os.Exit(1)
		}
	}
}
