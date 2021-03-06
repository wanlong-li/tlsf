package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
)

var (
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
)

func main() {
	infoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	warnLogger = log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime)
	errorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime)

	pNoVerifyServerCert := flag.Bool("no-verify", false, "skip verifying remote server cert")
	pCACert := flag.String("ca-cert", "", "CA certificate")
	pClientCert := flag.String("cert", "", "client certificate")
	pClientKey := flag.String("key", "", "client key")
	flag.Usage = func() {
		fmt.Println("Usage: tlsf [-no-verify] [-cacert ca_cert] [-cert client_cert] [-key client_key] remote_host:port bind_address:port")
		fmt.Println("\t-ca-cert: client CA certificate PEM file location (optional)")
		fmt.Println("\t-cert: client certificate PEM file location (optional)")
		fmt.Println("\t-key: client key PEM file location (optional)")
		fmt.Println("\t-no-verify: skip verifying server certificate (optional, default to false)")
	}
	flag.Parse()

	addresses := flag.Args()
	if len(addresses) != 2 {
		flag.Usage()
		os.Exit(2)
	}

	remoteAddr := addresses[0]
	bindAddr := addresses[1]
	tlsConfig, err := clientTLSConfig(*pNoVerifyServerCert, *pClientCert, *pClientKey, *pCACert)
	if err != nil {
		errorLogger.Println(err)
		os.Exit(-1)
	}

	listenAndDial(remoteAddr, bindAddr, tlsConfig)
}

func clientTLSConfig(skipVerifyServerCert bool, clientCert, clientKey, clientCACert string) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: skipVerifyServerCert,
	}

	if clientCert != "" || clientKey != "" || clientCACert != "" {
		cert, err := tls.LoadX509KeyPair(clientCert, clientKey)
		if err != nil {
			return nil, fmt.Errorf("error loading client certificate: %s", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}

		caCert, err := ioutil.ReadFile(clientCACert)
		if err != nil {
			warnLogger.Printf("CA certificate not loaded: %s\n", err)
		} else {
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tlsConfig.RootCAs = caCertPool
		}
	}
	return tlsConfig, nil
}

func listenAndDial(remoteAddr, localAddr string, tlsConfig *tls.Config) {
	listener, err := net.Listen("tcp", localAddr)

	if err != nil {
		errorLogger.Printf("[local] failed to listen: %s\n", err)
		return
	}

	infoLogger.Printf("[local] listening %s\n", localAddr)

	var connectionID uint64
	for {
		connectionID++
		lConn, err := listener.Accept()
		if err != nil {
			errorLogger.Printf("[local] failed to accept: %s\n", err)
			return
		}

		go forwardConnection(connectionID, lConn, remoteAddr, localAddr, tlsConfig)
	}
}

func forwardConnection(connectionID uint64, lConn net.Conn, remoteAddr, localAddr string, tlsConfig *tls.Config) {
	defer lConn.Close()

	rConn, err := dialRemote(remoteAddr, tlsConfig)
	if err != nil {
		errorLogger.Printf("[remote] failed to dial: %s\n", err)
		return
	}

	defer rConn.Close()

	infoLogger.Printf("[local] conn %d started\n", connectionID)
	defer func() {
		infoLogger.Printf("[local] conn %d ended\n", connectionID)
	}()

	go func() {
		pipe("remote", "local", rConn, lConn)
	}()
	pipe("local", "remote", lConn, rConn)
}

func dialRemote(remoteAddr string, tlsConfig *tls.Config) (net.Conn, error) {
	conn, err := tls.Dial("tcp", remoteAddr, tlsConfig)

	if err != nil {
		return nil, err
	}

	return conn, nil
}

func pipe(srcName, destName string, r io.ReadCloser, w io.WriteCloser) {
	for {
		byteWritten, err := io.Copy(w, r)
		if err != nil {
			if err != io.EOF {
				infoLogger.Printf("[%s->%s] failed to copy: %s\n", srcName, destName, err)
			}
			return
		}
		if byteWritten == 0 {
			infoLogger.Printf("[%s->%s] 0 byte\n", srcName, destName)
			return
		}
	}
}
