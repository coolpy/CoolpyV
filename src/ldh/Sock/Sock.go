package Sock

import (
	"net"
	"time"
	"net/url"
	"io"
	"fmt"
	"errors"
	"github.com/surgemq/message"
	"github.com/surgemq/surgemq/sessions"
	"sync/atomic"
)

var (
	ErrInvalidConnectionType error = errors.New("service: Invalid connection type")
	ErrInvalidSubscriber error = errors.New("service: Invalid subscriber")
	ErrBufferNotReady error = errors.New("service: buffer is not ready")
	ErrBufferInsufficientData error = errors.New("service: buffer has insufficient data.")
)
var (
	gsvcid uint64 = 0
)

const (
	DefaultKeepAlive = 300
	DefaultConnectTimeout = 2
	DefaultAckTimeout = 20
	DefaultTimeoutRetries = 3
	minKeepAlive = 30
)

type Server struct {
	KeepAlive      int
	ConnectTimeout int
	AckTimeout     int
	TimeoutRetries int
	quit           chan struct{}
	ln             net.Listener
	sessMgr          *sessions.Manager
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

//var headerPool = sync.Pool{
//	New:func() interface{} {
//		buf := make([]byte, 5)
//		return &buf
//	},
//}
//var bodyPool = sync.Pool{
//	New:func() interface{} {
//		buf := make([]byte, 128)
//		return &buf
//	},
//}

func (this *Server) handleConnection(c io.Closer) (svc *service, err error) {
	if c == nil {
		return nil, ErrInvalidConnectionType
	}

	defer func() {
		if err != nil {
			c.Close()
		}
	}()

	err = this.checkConfiguration()
	if err != nil {
		return nil, err
	}

	conn, ok := c.(net.Conn)
	if !ok {
		return nil, ErrInvalidConnectionType
	}

	conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(DefaultKeepAlive)))

	resp := message.NewConnackMessage()

	req, err := getConnectMessage(conn)
	if err != nil {
		if cerr, ok := err.(message.ConnackCode); ok {
			//glog.Debugf("request   message: %s\nresponse message: %s\nerror           : %v", mreq, resp, err)
			resp.SetReturnCode(cerr)
			resp.SetSessionPresent(false)
			writeMessage(conn, resp)
		}
		return nil, err
	}

	if string(req.Username()) != "username" && string(req.Password()) != "pwd" {
		resp.SetReturnCode(message.ErrBadUsernameOrPassword)
		resp.SetSessionPresent(false)
		writeMessage(conn, resp)
		return nil, err
	}

	if req.KeepAlive() == 0 {
		req.SetKeepAlive(minKeepAlive)
	}

	svc = &service{
		id:     atomic.AddUint64(&gsvcid, 1),
		client: false,

		keepAlive:      int(req.KeepAlive()),
		connectTimeout: this.ConnectTimeout,
		ackTimeout:     this.AckTimeout,
		timeoutRetries: this.TimeoutRetries,
		sessMgr:   this.sessMgr,
		conn:      conn,
	}

	//根据clientid提取缓存消息
	err = this.getSession(svc, req, resp)
	if err != nil {
		return nil, err
	}

	resp.SetReturnCode(message.ConnectionAccepted)

	if err = writeMessage(c, resp); err != nil {
		return nil, err
	}

	svc.inStat.increment(int64(req.Len()))
	svc.outStat.increment(int64(resp.Len()))

	if err := svc.start(); err != nil {
		svc.stop()
		return nil, err
	}

	glog.Infof("(%s) server/handleConnection: Connection established.", svc.cid())

	return svc, nil

	//header := headerPool.Get().(*[]byte)
	//defer headerPool.Put(header)
	//body := bodyPool.Get().(*[]byte)
	//defer bodyPool.Put(body)
	//for {
	//	if _, err := conn.Read(*header); err != nil {
	//		if err != io.EOF {
	//			fmt.Println("read error:", err)
	//		}
	//		break
	//	}
	//
	//	if _, err := mqtt.GetDefaultHeader((*header)); err == nil {
	//		if hd, err := mqtt.GetBufferHeader(*header); err == nil {
	//			(*body) = (*header)[hd.LenIndex:5]
	//			unreadlen := hd.Len - len(*body)
	//			if hd.Len <= 127 {
	//				if pl, err := conn.Read((*body)[5 - hd.LenIndex:hd.Len]); err != nil {
	//					fmt.Println(pl)
	//				}
	//			}
	//		}
	//	}
	//}
	//c.Close()
	//c = nil
	//return nil
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

func (this *Server) getSession(svc *service, req *message.ConnectMessage, resp *message.ConnackMessage) error {
	// If CleanSession is set to 0, the server MUST resume communications with the
	// client based on state from the current session, as identified by the client
	// identifier. If there is no session associated with the client identifier the
	// server must create a new session.
	//
	// If CleanSession is set to 1, the client and server must discard any previous
	// session and start a new one. This session lasts as long as the network c
	// onnection. State data associated with this session must not be reused in any
	// subsequent session.

	var err error

	// Check to see if the client supplied an ID, if not, generate one and set
	// clean session.
	if len(req.ClientId()) == 0 {
		req.SetClientId([]byte(fmt.Sprintf("internalclient%d", svc.id)))
		req.SetCleanSession(true)
	}

	cid := string(req.ClientId())

	// If CleanSession is NOT set, check the session store for existing session.
	// If found, return it.
	if !req.CleanSession() {
		if svc.sess, err = this.sessMgr.Get(cid); err == nil {
			resp.SetSessionPresent(true)

			if err := svc.sess.Update(req); err != nil {
				return err
			}
		}
	}

	// If CleanSession, or no existing session found, then create a new one
	if svc.sess == nil {
		if svc.sess, err = this.sessMgr.New(cid); err != nil {
			return err
		}

		resp.SetSessionPresent(false)

		if err := svc.sess.Init(req); err != nil {
			return err
		}
	}

	return nil
}

func (this *service) publish(msg *message.PublishMessage, onComplete OnCompleteFunc) error {
	//glog.Debugf("service/publish: Publishing %s", msg)
	_, err := this.writeMessage(msg)
	if err != nil {
		return fmt.Errorf("(%s) Error sending %s message: %v", this.cid(), msg.Name(), err)
	}

	switch msg.QoS() {
	case message.QosAtMostOnce:
		if onComplete != nil {
			return onComplete(msg, nil, nil)
		}

		return nil

	case message.QosAtLeastOnce:
		return this.sess.Pub1ack.Wait(msg, onComplete)

	case message.QosExactlyOnce:
		return this.sess.Pub2out.Wait(msg, onComplete)
	}

	return nil
}