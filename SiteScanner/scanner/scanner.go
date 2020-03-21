package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	dbPath    = "params"
	heuristic = 0
)

func Requester(url string) ([]string, error) {
	urlToCheck := []string{}

	wg := sync.WaitGroup{}
	var mutex = &sync.Mutex{}

	fmt.Println(url)

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
	wg := sync.WaitGroup{}
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	log.Println(fileName, " Starting...")

	file, err := os.Open(fileName)
	if err != nil {
		log.Println(err)
		return
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		wg.Add(1)
		getParametr := scanner.Text()

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

		defer resp.Body.Close()
		go func(resp *http.Response) {
			respBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println(err)
			}

			respLen := float64(len(respBody))
			if resp.Status == "200" && (respLen < float64(heuristic)*0.75 || respLen > float64(heuristic)*1.25) {
				mu.Lock()
				urlToCheck = append(urlToCheck, url)
				mu.Unlock()
			}

			wg.Done()
		}(resp)

		q.Del(getParametr)
	}

	wg.Wait()
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
	url, _, err := reader.ReadLine()
	if err != nil {
		log.Fatal(err)
	}

	err = Heuristic(string(url))
	if err != nil {
		fmt.Println(err)
	} else {
		urlToCheck, err := Requester(string(url))
		if err != nil {
			fmt.Println(err)
			return
		}
		for url := range urlToCheck {
			fmt.Println("[+] Possible hidden get parametr - ", url)
		}
	}
}
