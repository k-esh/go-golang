package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var downloadDir = "/tmp/image_down"
const NUM_WORKERS_DFLT = 50

var myWaitGroup sync.WaitGroup

type downloadFail struct {
	shortDesc string
	resource  string
	code      int
	details   string
}

func init() {
	log.SetOutput(os.Stdout) //stdout not stderr
}

//main drives the process -
// takes one argument for the input file with URLs to download
// uses a channel to feed go routine which does the download concurrently
func main() {

	// process args
	// 1 - input file with the urls
	// 2 - number of concurrent go routines (can take default)
	var inputFile string
	if len(os.Args) > 1 {
		inputFile = os.Args[1]
	} else {
		log.Fatal("argument is missing - which input file has URLs?")
	}
	var numWorkers int
	if len(os.Args) <= 2 {
		numWorkers = NUM_WORKERS_DFLT
		log.Println("number of workers [takes default]: ", numWorkers)
	} else {
		numWorkers, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatal("number of workers argument must be numeric: ", os.Args[2])
		}
		log.Println("number of workers: ", numWorkers)
	}


	urls, err := getUrls(inputFile) //read URL file into a slice
	if err != nil {
		log.Fatal(err)
	}

	prepDownloadDir() //create dir for downloading into
	log.Println("input file name: ", inputFile)
	log.Println("count of urls to download: ", len(urls))
	log.Println("prepped download folder: ", downloadDir)

	urlStream := make(chan string, len(urls))
	for _, url := range urls {
		urlStream <- url //put every URL string into the channel
	}
	close(urlStream) //close the channel - nothing more will be written to it


	downFailStream := make(chan downloadFail, len(urls)) //workers will write errors here

	myWaitGroup.Add(numWorkers) //tell waitgroup about the Number of Workers
	start := time.Now() //stopwatch begins

	log.Println("processing begins - number of CPUs: ", runtime.NumCPU())
	for w := 1; w <= numWorkers; w++ {
		go downloadFile(urlStream, downFailStream) //start go routines up to total number of workers
	}

	myWaitGroup.Wait() //block until all workers are done - when each reads channel as empty

	elapsed := time.Since(start).Minutes()
	log.Println("download time elapsed minutes", elapsed)

	handleErrorsList(downFailStream)

	log.Println("process is complete")
}

//handle any errors that were put in the error channel
func handleErrorsList(downFailStream chan downloadFail) {

	close(downFailStream)
	log.Println("download error count:", len(downFailStream))

	if len(downFailStream) > 0 {
		errorListFile, err := os.Create(downloadDir + "/download_fail.txt")
		if err != nil {
			log.Fatal(err)
		}
		buff := bufio.NewWriter(errorListFile)

		for errorElement := range downFailStream {
			buff.WriteString(fmt.Sprintf("%#v\n", errorElement))
		}
		buff.Flush()
		log.Println("see error file for details: ", errorListFile)
	}
}

// extract rightmost part of a URL, and prepend with a directory name. Useful
// to take a URL (like an image), and choose where to download to.
//  e.g. input 'https://yahoo.com/marissa.jpg' yields '/tmp/foo/marissa.jpg'
func fileFromURL(url string) string {
	lastSlash := strings.LastIndex(url, "/")
	target := downloadDir + "/" + url[lastSlash+1:]
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

// drop then recreate the directory to hold downloads
func prepDownloadDir() {
	err := os.RemoveAll(downloadDir) //whack download dir
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll(downloadDir, os.ModePerm) //create download dir
	if err != nil {
		log.Fatal(err)
	}
}

// read URL from a channel, and download from URL to disk
func downloadFile(urlStream chan string, downFailStream chan downloadFail) {
	defer myWaitGroup.Done()

	//forever loop - until channel read comes up empty
	for {
		url, ok := <-urlStream
		if !ok {
			return // channel is empty and closed, worker to end
		}

		out, err := os.Create(fileFromURL(url))
		if err != nil {
			log.Fatal(err) //die when cannot write to disk
		}

		response, httpError := http.Get(url)

		if httpError == nil {
			_, err = io.Copy(out, response.Body) // write the body to file

			if err != nil {
				log.Fatal("IO Error saving file", err.Error(), url)
			} else {
				if response.StatusCode != http.StatusOK { // not 200 OK
					downFailStream <- downloadFail{
						shortDesc: "NOTOK",
						resource:  url,
						code:      response.StatusCode,
						details:   ""}
				}
			}
			out.Close()
			response.Body.Close()
		} else {
			downFailStream <- downloadFail{
				shortDesc: "HTTPERROR",
				resource:  url,
				code:      -1,
				details:   httpError.Error()}
		}
	}
}
