package main

import (
	"bufio"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

var (
	RepeaterClient *http.Client
	historyPath    = "../history/"
)

func History(writer http.ResponseWriter, req *http.Request) {
	files, err := ioutil.ReadDir(historyPath)
	if err != nil {
		log.Println(err)
	}

	for _, file := range files {
		fileDesc, err := os.Open(historyPath + file.Name())
		if err != nil {
			log.Println(err)
		}

		reader := bufio.NewReader(fileDesc)

		line, err := reader.ReadBytes('\n')
		if err != nil {
			log.Println(err)
		}

		writer.Write(line)
		writer.Write([]byte(file.Name() + "\n"))
	}
}

func Resend(writer http.ResponseWriter, req *http.Request) {
	reqID, ok := mux.Vars(req)["id"]
	if !ok {
		log.Println("No such element")
		return
	}

	file, err := os.Open(historyPath + reqID)
	if err != nil {
		log.Println("No such file")
		return
	}

	buf := bufio.NewReader(file)

	repReq, err := http.ReadRequest(buf)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(repReq.Host)
	log.Println(repReq.URL.RawPath)
	log.Println(repReq.URL.RawQuery)

	request, err := http.NewRequest(repReq.Method, "http://"+repReq.Host, repReq.Body)
	if err != nil {
		log.Println(err)
		return
	}
	response, err := RepeaterClient.Do(request)
	if err != nil {
		log.Println(err)
		return
	}

	CopyHeader(writer.Header(), response.Header)
	writer.WriteHeader(response.StatusCode)
	io.Copy(writer, response.Body)
}

func CopyHeader(dst, src http.Header) {
	for k, vv := range src {
		var value string
		for _, v := range vv {
			value = value + v
			dst.Add(k, v)
		}
	}
}

func main() {
	r := mux.NewRouter()
	RepeaterClient = &http.Client{}
	r.HandleFunc("/history", History).Methods("GET")
	r.HandleFunc("/request/{id}", Resend).Methods("GET")

	err := http.ListenAndServe(":8090", r)
	if err != nil {
		log.Fatal(err)
		return
	}
}
