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
	// "time" 	// "errors"
)

func init() {
	log.SetOutput(os.Stdout) //stdout not stderr
}

var wg sync.WaitGroup

func fileFromUrl(url string) string {
	file := "/tmp/gofiles/"
	lastSlash := strings.LastIndex(url, "/")
	file += url[lastSlash+1 : len(url)]
	return file
}

func responseTime(channelU chan string) {
	defer wg.Done()

	for {
		url, ok := <-channelU
		if !ok {
			return // channels is empty and closed, worker to end
		}

		out, err := os.Create(fileFromUrl(url))
		if err != nil {
			log.Fatal(err)
		}

		// start := time.Now()
		res, httpError := http.Get(url)
		if httpError == nil {
			_, err = io.Copy(out, res.Body) // Write the body to file
			if err != nil {
				fmt.Println(err) // log.Fatal(err)
			}
			out.Close()
			res.Body.Close()
		} else {
			log.Println(httpError) // log.Fatal(err) or os.Exit(1)
		}
		// elapsed := time.Since(start).Seconds()
		// log.Println("done with file name: ", savedFile)
		// log.Println("%s took %v seconds \n", url, elapsed)
	}
}

func main() {
	log.Println("function main begins")
	workers := 110

	var inputFile string
	if len(os.Args) > 1 {
		inputFile = os.Args[1]
		log.Println("input file name: ", inputFile)
	} else {
		log.Fatal("argument required - which input file?")
	}

	urls, err := getUrls(inputFile)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("count of urls: ", len(urls))
	wg.Add(workers)
	channelU := make(chan string, 500)

	for w := 1; w <= workers; w++ {
		go responseTime(channelU)
	}

	for _, url := range urls {
		channelU <- url
	}
	close(channelU)

	wg.Wait()
  log.Println("process is comlete")
}

// func getUrls(inFile string) ([]*string, error) {
//   log.Println("function GetUrls begins with:", inFile)
//    return nil, errors.New("I am an error")
// }

// readLines reads a whole file into memory
// and returns a slice of its lines.
func getUrls(path string) ([]string, error) {
	log.Println("getting the URLs from file: ", path)

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
