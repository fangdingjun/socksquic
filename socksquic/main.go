package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	socks "github.com/fangdingjun/socks-go"
	"github.com/golang/glog"
	quic "github.com/lucas-clemente/quic-go"
)

type streamConn struct {
	quic.Stream
	quic.Session
}

func (sc streamConn) Close() error {
	return sc.Stream.Close()
}

// handle QUIC stream
func handleStream(stream quic.Stream, sess quic.Session) {
	s := streamConn{Stream: stream, Session: sess}

	// serve socks request
	s1 := socks.Conn{Conn: s}
	s1.Serve()
}

// handle QUIC session
func handleSession(sess quic.Session) {
	glog.Infof("new session from %s", sess.RemoteAddr())

	defer sess.Close(nil)

	for {
		stream, err := sess.AcceptStream()
		if err != nil {
			glog.Error(err)
			break
		}
		go handleStream(stream, sess)
	}
}

func main() {
	var port int
	var cert, keyfile string
	flag.IntVar(&port, "port", 443, "listen port")
	flag.StringVar(&cert, "cert", "server.pem", "server certificate file")
	flag.StringVar(&keyfile, "key", "private.pem", "server private key file")
	flag.Parse()

	certs, err := tls.LoadX509KeyPair(cert, keyfile)
	if err != nil {
		glog.Fatal(err)
	}

	// listen for QUIC
	l, err := quic.ListenAddr(
		fmt.Sprintf(":%d", port),
		&tls.Config{
			Certificates: []tls.Certificate{certs},
		},
		nil)
	if err != nil {
		glog.Fatal(err)
	}

	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			glog.Error(err)
			break
		}
		go handleSession(c)
	}
}
