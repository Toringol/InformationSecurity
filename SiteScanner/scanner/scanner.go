package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	dbPath    = "params.txt"
	heuristic = 0
)

func Requester(url string) ([]string, error) {
	urlToCheck := []string{}
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	fmt.Println(url)

	file, err := os.Open(dbPath)
	if err != nil {
		return urlToCheck, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		getParametr := scanner.Text()

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return urlToCheck, err
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
		resp_body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}

		respLen := float64(len(resp_body))
		if resp.Status == "200" && (respLen < float64(heuristic)*0.75 || respLen > float64(heuristic)*1.25) {
			urlToCheck = append(urlToCheck, url)
		}

		q.Del(getParametr)
	}

	return urlToCheck, nil
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
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	heuristic = len(resp_body)

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
