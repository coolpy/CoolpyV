package Sock

import (
	"net"
	"time"
	"net/url"
	"io"
	"fmt"
	"sync"
	"mqtt"
)

const (
	DefaultKeepAlive = 300
	DefaultConnectTimeout = 2
	DefaultAckTimeout = 20
	DefaultTimeoutRetries = 3
)

type Server struct {
	KeepAlive      int
	ConnectTimeout int
	AckTimeout     int
	TimeoutRetries int
	quit           chan struct{}
	ln             net.Listener
}

func (this *Server) ListenAndServe(uri string) error {
	this.quit = make(chan struct{})

	u, err := url.Parse(uri)
	if err != nil {
		return err
	}

	this.ln, err = net.Listen(u.Scheme, u.Host)
	if err != nil {
		return err
	}
	defer this.ln.Close()

	fmt.Println("server/ListenAndServe: server is ready...")

	var tempDelay time.Duration // how long to sleep on accept failure

	for {
		conn, err := this.ln.Accept()

		if err != nil {
			// http://zhen.org/blog/graceful-shutdown-of-go-net-dot-listeners/
			select {
			case <-this.quit:
				return nil
			default:
			}

			// Borrowed from go1.3.3/src/pkg/net/http/server.go:1699
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				fmt.Println("server/ListenAndServe: Accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return err
		}

		go this.handleConnection(conn)
	}
}

var headerPool = sync.Pool{
	New:func() interface{} {
		buf := make([]byte, 2)
		return &buf
	},
}
var LenPool = sync.Pool{
	New:func() interface{} {
		buf := make([]byte, 4)
		return &buf
	},
}
var CheckLenPool = sync.Pool{
	New:func() interface{} {
		buf := new(byte)
		return &buf
	},
}

func (this *Server) handleConnection(c io.Closer) error {
	defer c.Close()

	this.checkConfiguration()

	conn, ok := c.(net.Conn)
	if !ok {
		return nil
	}

	conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(DefaultKeepAlive)))

	header := headerPool.Get().(*[]byte)
	len := LenPool.Get().(*[]byte)
	clen := CheckLenPool.Get().(*byte)
	defer headerPool.Put(header)
	defer LenPool.Put(len)
	defer CheckLenPool.Put(clen)
	for {
		if _, err := conn.Read(*header); err != nil {
			if err != io.EOF {
				fmt.Println("read error:", err)
			}
			break
		}

		mt,_ := mqtt.GetDefaultHeader((*header))
		if (mt.IsContinue == mqtt.Continue) {
			*len[0] = header[1]
			for i := 1; i < 4; i++ {
				if _, err := conn.Read(*clen); err == nil {
					*len[i] = *clen
					if !mqtt.CheckIsContinue((*clen)) {
                                          break
					}
				}
			}
		}

		pkglen,_ := mqtt.GetBufferHeader(*len)
		fmt.Println(pkglen)

		//if read_len == 0 {
		//	break // connection already closed by client
		//} else if string(*header) == "ti" {
		//	daytime := strconv.FormatInt(time.Now().Unix(), 10)
		//	conn.Write([]byte(daytime))
		//} else {
		//	daytime := time.Now().String()
		//	conn.Write([]byte(daytime))
		//}
	}

	return nil
}

func (this *Server) checkConfiguration() {
	if this.KeepAlive == 0 {
		this.KeepAlive = DefaultKeepAlive
	}

	if this.ConnectTimeout == 0 {
		this.ConnectTimeout = DefaultConnectTimeout
	}

	if this.AckTimeout == 0 {
		this.AckTimeout = DefaultAckTimeout
	}

	if this.TimeoutRetries == 0 {
		this.TimeoutRetries = DefaultTimeoutRetries
	}
}