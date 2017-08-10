package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/golang/glog"
	quic "github.com/lucas-clemente/quic-go"
	"io"
	"net"
)

type quicConn struct {
	srv        string
	tlsConfig  *tls.Config
	quicConfig *quic.Config
	remote     net.Addr
}

func (qc *quicConn) handleConn(c net.Conn) {
	glog.Infof("new connection from: %s", c.RemoteAddr())

	defer c.Close()

	c1, err := net.ListenPacket("udp", ":0")
	if err != nil {
		glog.Error(err)
		return
	}

	sess, err := quic.Dial(c1, qc.remote, qc.srv, qc.tlsConfig, qc.quicConfig)
	if err != nil {
		glog.Error(err)
		return
	}
	defer sess.Close(nil)

	stream, err := sess.OpenStreamSync()
	if err != nil {
		glog.Error(err)
		return
	}

	defer stream.Close()

	forward(c, stream)
}

func forward(c, c1 io.ReadWriter) {
	ch := make(chan struct{}, 2)

	go func() {
		io.Copy(c, c1)
		ch <- struct{}{}
	}()

	go func() {
		io.Copy(c1, c)
		ch <- struct{}{}
	}()

	<-ch
}

func main() {
	var port int
	var err error
	var server string

	flag.IntVar(&port, "port", 8081, "port to listen")
	flag.StringVar(&server, "server", "", "server to connect to")
	flag.Parse()

	raddr, err := net.ResolveUDPAddr("udp", server)
	if err != nil {
		glog.Fatal(err)
	}

	glog.Infof("forward to %s", server)

	remoteConn := &quicConn{
		srv:        server,
		tlsConfig:  &tls.Config{InsecureSkipVerify: true},
		quicConfig: &quic.Config{KeepAlive: true},
		remote:     raddr,
	}

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
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
		go remoteConn.handleConn(c)
	}
}
