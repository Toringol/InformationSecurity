package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	startURL  = ""
	dbPath    = "../params"
	heuristic = 0
)

func Requester(url string) ([]string, error) {
	urlToCheck := []string{}

	wg := sync.WaitGroup{}
	var mutex = &sync.Mutex{}

	files, err := ioutil.ReadDir(dbPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		wg.Add(1)
		filePath := dbPath + "/" + file.Name()
		go func(filePath string) {
			Process(filePath, url, urlToCheck, mutex)
			wg.Done()
		}(filePath)
	}

	wg.Wait()

	return urlToCheck, nil
}

func Process(fileName string, url string, urlToCheck []string, mu *sync.Mutex) {

	log.Println(fileName, " Starting...")

	file, err := os.Open(fileName)
	if err != nil {
		log.Println(err)
		return
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		getParametr := scanner.Text()

		go func(mu *sync.Mutex) {
			fetch(url, getParametr, urlToCheck, mu)
		}(mu)

	}

}

func fetch(url string, getParametr string, urlsToCheck []string, mu *sync.Mutex) {
	start := time.Now()
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Add("Accept", "application/json")

	q := req.URL.Query()
	q.Add(getParametr, "1")
	req.URL.RawQuery = q.Encode()

	fmt.Println(req.URL.String())

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Errored when sending request to the server")
	}

	nbytes, err := io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Println(err)
		return
	}

	secs := time.Since(start).Seconds()

	if float64(nbytes) < float64(heuristic)*0.75 || float64(nbytes) > float64(heuristic)*1.25 {
		fmt.Sprintf("%.2fs  %7d  %s", secs, nbytes, url)
		mu.Lock()
		urlsToCheck = append(urlsToCheck, url)
		mu.Unlock()
	}
}

func Heuristic(url string) error {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	heuristic = len(respBody)

	return nil
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("What site you want to test: ")
	startURL, _, err := reader.ReadLine()
	if err != nil {
		log.Fatal(err)
	}

	err = Heuristic(string(startURL))
	if err != nil {
		fmt.Println(err)
		return
	}

	urlToCheck, err := Requester(string(startURL))
	if err != nil {
		fmt.Println(err)
		return
	}

	for url := range urlToCheck {
		fmt.Println("[+] Possible hidden get parametr - ", url)
	}
}
