// Copyright © 2019 Moises P. Sena <moisespsena@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/gorilla/websocket"

	"github.com/moisespsena-go/logging"
)

var log = logging.NewLogger(name)

type Chan struct {
	c chan []byte
}

func (this *Chan) Read(p []byte) (n int, err error) {
	d := <-this.c
	if d == nil || len(d) == 0 {
		err = io.EOF
		return
	}
	n = copy(p, d)
	return
}

func (this *Chan) Write(p []byte) (n int, err error) {
	this.c <- p
	return len(p), nil
}

// maxBufferSize specifies the size of the buffers that
// are used to temporarily hold data from the UDP packets
// that we receive.
const maxBufferSize = 1024

type UDPServerReader struct {
	proto, address string
	w              io.Writer
	buffer         []byte
	pc             net.PacketConn
}

func (this *UDPServerReader) Close() error {
	if this.pc != nil {
		return this.pc.Close()
	}
	return nil
}

func NewUDPServer(addr string, w io.Writer) *UDPServerReader {
	proto, addr := parseAddr(addr)
	return &UDPServerReader{proto: proto, address: addr, w: w, buffer: make([]byte, maxBufferSize)}
}

func (this *UDPServerReader) ListenAndServe() (err error) {
	this.pc, err = net.ListenPacket(this.proto, this.address)
	if err != nil {
		return
	}

	defer this.pc.Close()

	var n int

	for {
		if n, _, err = this.pc.ReadFrom(this.buffer); err != nil {
			return
		}

		if n > 0 {
			b := this.buffer[0:n]
			if b[n-1] != '\n' {
				b = append(b, '\n')
			}
			if _, err = this.w.Write(b); err != nil {
				return
			}
		}
	}
}

type HTTPServerReader struct {
	proto, addr string
	s           *http.Server
	l           net.Listener
	w           io.Writer
}

func NewHTTPServerReader(addr string, w io.Writer) *HTTPServerReader {
	proto, addr := parseAddr(addr)
	proto = "tcp" + strings.TrimPrefix(proto, "http")
	return &HTTPServerReader{proto: proto, addr: addr, w: w}
}

func (this *HTTPServerReader) Close() error {
	this.s.Shutdown(context.Background())
	return nil
}

func (this *HTTPServerReader) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if websocket.IsWebSocketUpgrade(r) {
		var upgrader = websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		for {
			_, b, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if b[len(b)] != '\n' {
				b = append(b, '\n')
			}
			if _, err := this.w.Write(b); err != nil {
				return
			}

		}
	} else if r.Method == http.MethodPost {
		if _, err := io.Copy(this.w, &lfreader{r: r.Body}); err != nil {
			log.Error("TCP copy failed: " + err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	} else {
		http.Error(w, "bad http method", http.StatusBadRequest)
	}
}

func (this *HTTPServerReader) ListenAndServe() (err error) {
	if this.l, err = net.Listen(this.proto, this.addr); err != nil {
		return
	}
	this.s = &http.Server{Handler: this}
	return this.s.Serve(&tcpKeepAliveListener{this.l.(*net.TCPListener)})
}

type TCPServerReader struct {
	proto, addr string
	l           net.Listener
	w           io.Writer
}

func (this *TCPServerReader) Close() error {
	return this.l.Close()
}

func NewTCPServerReader(addr string, w io.Writer) *TCPServerReader {
	proto, addr := parseAddr(addr)
	return &TCPServerReader{proto: proto, addr: addr, w: w}
}

func (this *TCPServerReader) ListenAndServe() (err error) {
	if this.l, err = net.Listen(this.proto, this.addr); err != nil {
		return
	}

	this.l = &tcpKeepAliveListener{this.l.(*net.TCPListener)}

	var c net.Conn
	for {
		if c, err = this.l.Accept(); err != nil {
			return
		}
		go func(c net.Conn) {
			if _, err := io.Copy(this.w, c); err != nil && err != io.EOF {
				log.Errorf("TCP copy failed: %s", err.Error())
			}
		}(c)
	}
}

type lfreader struct {
	r    io.Reader
	done bool
}

func (this lfreader) Read(p []byte) (n int, err error) {
	if this.done {
		err = io.EOF
	} else if n, err = this.r.Read(p); err == io.EOF {
		if n < len(p) {
			p[n] = '\n'
			n++
		} else {
			p[0] = '\n'
			n = 1
			err = nil
		}
		this.done = true
	}
	return
}

func parseAddr(address string) (network, addr string) {
	pos := strings.IndexRune(address, ':')
	network, addr = address[0:pos], address[pos+1:]
	if !strings.HasSuffix(network, "6") && strings.IndexRune(addr, '[') >= 0 {
		network += "6"
	}
	return
}

func isProto(address, proto string) bool {
	pos := strings.IndexRune(address, ':')
	for unicode.IsDigit(rune(address[pos-1])) {
		pos--
	}
	part := address[0:pos]
	return proto == part
}

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}
