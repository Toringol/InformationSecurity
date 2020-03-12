package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/base32"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

var (
	historyPath = "history/"
)

var (
	pemPath = flag.String("pem", "server.pem", "Path to pem file")
	keyPath = flag.String("key", "server.key", "Path to key file")
	proto   = flag.String("proto", "https", "Proxy protocol (http or https)")
)

type Opts struct {
	PemPath string
	KeyPath string
	Proto   string
}

func Store(req *http.Request) (err error) {
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
	log.Println(string(requestLine), strHash)

	fileName := strHash
	if _, err = os.Stat(fileName); os.IsNotExist(err) {
		var file *os.File
		file, err = os.Create(historyPath + fileName)
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

func handleTunneling(w http.ResponseWriter, r *http.Request) {
	dest_conn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	client_conn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}

	go transfer(dest_conn, client_conn)
	go transfer(client_conn, dest_conn)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

func handleHTTP(w http.ResponseWriter, req *http.Request) {
	log.Println("HTTP: ", req.Header)
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	defer resp.Body.Close()

	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func main() {
	flag.Parse()

	opts := Opts{
		PemPath: *pemPath,
		KeyPath: *keyPath,
		Proto:   *proto,
	}

	if opts.Proto != "http" && opts.Proto != "https" {
		log.Fatal("Protocol must be either http or https")
	}

	server := &http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			err := Store(r)
			if err != nil {
				log.Println("Unable to store round trip result, err:", err)
			}

			if r.Method == http.MethodConnect {
				handleTunneling(w, r)
			} else {
				handleHTTP(w, r)
			}
		}),
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	if opts.Proto == "http" {
		log.Fatal(server.ListenAndServe())
	} else {
		log.Fatal(server.ListenAndServeTLS(opts.PemPath, opts.KeyPath))
	}
}
