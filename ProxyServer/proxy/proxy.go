package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/base32"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/Toringol/InformationSecurity/ProxyServer/certificates"
)

var (
	historyPath = "../history/"
)

var rootCertificate certificates.Cert

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
	cert, err := certificates.CreateLeafCertificate(r.Host)
	if err != nil {
		log.Println(err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*cert},
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return certificates.CreateLeafCertificate(info.ServerName)
		},
	}

	dest_conn, err := tls.Dial("tcp", r.Host, tlsConfig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	client_conn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	_, err = client_conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

	tlsConnection := tls.Server(client_conn, tlsConfig)
	err = tlsConnection.Handshake()

	go transfer(dest_conn, tlsConnection, true, r.URL.Host)
	go transfer(tlsConnection, dest_conn, false, r.URL.Host)
}

func transfer(destination io.WriteCloser, source io.ReadCloser, copy bool, requestHost string) {
	defer destination.Close()
	defer source.Close()
	if copy {
		buffer := &bytes.Buffer{}
		duplicateSources := io.MultiWriter(destination, buffer) //we copy data from source into buffer and destination
		io.Copy(duplicateSources, source)
	} else {
		io.Copy(destination, source)
	}
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

	rootCertificate = certificates.GetRootCertificate()

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
		return
	}

}
