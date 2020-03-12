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
	dbPath = "params.txt"
)

func Requester(url string) ([]string, error) {
	urlToCheck := []string{}
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	log.Println(url)

	file, err := os.Open(dbPath)
	if err != nil {
		return urlToCheck, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		getParametr := scanner.Text()

		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Add("Accept", "application/json")

		q := req.URL.Query()
		q.Add(getParametr, "1")
		req.URL.RawQuery = q.Encode()

		log.Println(req.URL.String())

		resp, err := client.Do(req)

		if err != nil {
			fmt.Println("Errored when sending request to the server")
		}

		defer resp.Body.Close()
		resp_body, _ := ioutil.ReadAll(resp.Body)

		fmt.Println(resp.Status)
		fmt.Println(string(resp_body))

		q.Del(getParametr)
	}

	return urlToCheck, nil
}

func main() {

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("What site you want to test: ")
	url, _, err := reader.ReadLine()
	if err != nil {
		log.Fatal(err)
	}

	log.Println(url)

	Requester(string(url))
}
