package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func helloServer(w http.ResponseWriter, req *http.Request) {
	time.Sleep(3 * time.Second) // long connection
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Hello.\n"))
}

func main() {

	caCert, err := ioutil.ReadFile("cert.pem")
	if err != nil {
		log.Fatal("fail to load CA cert: ", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}

	handler := http.NewServeMux()
	handler.HandleFunc("/", helloServer)

	server := &http.Server{
		Handler:   handler,
		Addr:      "localhost:8443",
		TLSConfig: tlsConfig,
	}

	err = server.ListenAndServeTLS("cert.pem", "key.pem")
	if err != nil {
		log.Fatal("failed to start server: ", err)
	}
}
