package storage

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/base32"
	"log"
	"net/http"
	"os"
	"strconv"
)

var storingPath = "/tmp/"

func init() {
	err := os.MkdirAll(storingPath, os.ModeDir|os.ModePerm)
	if err != nil && err != os.ErrExist {
		log.Fatal("Unable to create directory by storing path", err)
	}
	err = os.Chdir(storingPath)
	if err != nil {
		log.Fatal("Unable to change working directory to storing path, err:", err)
	}
}

func Store(req *http.Request, resp *http.Response) (err error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	err = req.Write(buf)
	if err != nil {
		log.Println("Unable to write request to buffer, err:", err)
		return
	}

	hash := md5.Sum(buf.Bytes())
	strHash := base32.StdEncoding.EncodeToString(hash[:])
	var requestLine []byte
	reqBuf := bufio.NewReader(buf)
	requestLine, _, err = reqBuf.ReadLine()
	if err != nil {
		log.Println("Unable to read request line from buffer, err:", err)
		return
	}
	log.Println(resp.StatusCode, string(requestLine), strHash)

	fileName := strHash + strconv.Itoa(resp.StatusCode)
	if _, err = os.Stat(fileName); os.IsNotExist(err) {
		var file *os.File
		file, err = os.Create(fileName)
		if err != nil {
			log.Println("Unable to create new storing file, err:", err)
			return
		}
		defer file.Close()

		_, err = file.Write(append(append(requestLine, byte('\r')), byte('\n')))
		if err != nil {
			log.Println("Unable to write request line to file, err", err)
			return
		}

		_, err = reqBuf.WriteTo(file)
		if err != nil {
			log.Println("Unable to write buffer to file, error:", err)
			return
		}
	} else {
		//Storing already exist (err = nil) or it is unexpected error
		return
	}
	return nil
}
