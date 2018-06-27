package main

import (
  "log"
  "os"
  // "errors"
  "bufio"
)
func init() {
  log.SetOutput(os.Stdout) //stdout not stderr
}

func main() {
  log.Println("function main begins")

  var inputFile string
  if len(os.Args) > 1 {
    inputFile = os.Args[1]
  } else {
    log.Fatal("argument required - which input file?")
  }

  log.Println("input file name: ", inputFile)
  urls, err := getUrls(inputFile)
  if err != nil {
    log.Fatal(err)
  }
  log.Println("count of urls: ", len(urls))

}

  // func getUrls(inFile string) ([]*string, error) {
  //   log.Println("function GetUrls begins with:", inFile)
  //    return nil, errors.New("I am an error")
  // }

  // readLines reads a whole file into memory
// and returns a slice of its lines.
func getUrls(path string) ([]string, error) {
  log.Println("function GetUrls begins with:", path)

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
