package Sock

import (
	"sync/atomic"
	"io"
	"sync"
	"github.com/surgemq/message"
	"github.com/surgemq/surgemq/sessions"
)

type stat struct {
	bytes int64
	msgs  int64
}

func (this *stat) increment(n int64) {
	atomic.AddInt64(&this.bytes, n)
	atomic.AddInt64(&this.msgs, 1)
}

type (
	OnCompleteFunc func(msg, ack message.Message, err error) error
	OnPublishFunc  func(msg *message.PublishMessage) error
)

type service struct {
	id uint64
	client bool
	keepAlive int
	connectTimeout int
	ackTimeout int
	timeoutRetries int
	conn io.Closer
	wmu sync.Mutex
	closed int64
	done chan struct{}
	inStat  stat
	outStat stat
	// Incoming data buffer. Bytes are read from the connection and put in here.
	in *buffer

	// Outgoing data buffer. Bytes written here are in turn written out to the connection.
	out *buffer

	// onpub is the method that gets added to the topic subscribers list by the
	// processSubscribe() method. When the server finishes the ack cycle for a
	// PUBLISH message, it will call the subscriber, which is this method.
	//
	// For the server, when this method is called, it means there's a message that
	// should be published to the client on the other end of this connection. So we
	// will call publish() to send the message.
	onpub OnPublishFunc
	sessMgr *sessions.Manager
	sess *sessions.Session
}

func (this *service) start() error {
	var err error

	// Create the incoming ring buffer
	this.in, err = newBuffer(defaultBufferSize)
	if err != nil {
		return err
	}

	// Create the outgoing ring buffer
	this.out, err = newBuffer(defaultBufferSize)
	if err != nil {
		return err
	}

	// If this is a server
	if !this.client {
		// Creat the onPublishFunc so it can be used for published messages
		this.onpub = func(msg *message.PublishMessage) error {
			if err := this.publish(msg, nil); err != nil {
				glog.Errorf("service/onPublish: Error publishing message: %v", err)
				return err
			}

			return nil
		}

		// If this is a recovered session, then add any topics it subscribed before
		topics, qoss, err := this.sess.Topics()
		if err != nil {
			return err
		} else {
			for i, t := range topics {
				this.topicsMgr.Subscribe([]byte(t), qoss[i], &this.onpub)
			}
		}
	}

	// Processor is responsible for reading messages out of the buffer and processing
	// them accordingly.
	this.wgStarted.Add(1)
	this.wgStopped.Add(1)
	go this.processor()

	// Receiver is responsible for reading from the connection and putting data into
	// a buffer.
	this.wgStarted.Add(1)
	this.wgStopped.Add(1)
	go this.receiver()

	// Sender is responsible for writing data in the buffer into the connection.
	this.wgStarted.Add(1)
	this.wgStopped.Add(1)
	go this.sender()

	// Wait for all the goroutines to start before returning
	this.wgStarted.Wait()

	return nil
}

// FIXME: The order of closing here causes panic sometimes. For example, if receiver
// calls this, and closes the buffers, somehow it causes buffer.go:476 to panid.
func (this *service) stop() {
	defer func() {
		// Let's recover from panic
		if r := recover(); r != nil {
			glog.Errorf("(%s) Recovering from panic: %v", this.cid(), r)
		}
	}()

	doit := atomic.CompareAndSwapInt64(&this.closed, 0, 1)
	if !doit {
		return
	}

	// Close quit channel, effectively telling all the goroutines it's time to quit
	if this.done != nil {
		glog.Debugf("(%s) closing this.done", this.cid())
		close(this.done)
	}

	// Close the network connection
	if this.conn != nil {
		glog.Debugf("(%s) closing this.conn", this.cid())
		this.conn.Close()
	}

	this.in.Close()
	this.out.Close()

	// Wait for all the goroutines to stop.
	this.wgStopped.Wait()

	glog.Debugf("(%s) Received %d bytes in %d messages.", this.cid(), this.inStat.bytes, this.inStat.msgs)
	glog.Debugf("(%s) Sent %d bytes in %d messages.", this.cid(), this.outStat.bytes, this.outStat.msgs)

	// Unsubscribe from all the topics for this client, only for the server side though
	if !this.client && this.sess != nil {
		topics, _, err := this.sess.Topics()
		if err != nil {
			glog.Errorf("(%s/%d): %v", this.cid(), this.id, err)
		} else {
			for _, t := range topics {
				if err := this.topicsMgr.Unsubscribe([]byte(t), &this.onpub); err != nil {
					glog.Errorf("(%s): Error unsubscribing topic %q: %v", this.cid(), t, err)
				}
			}
		}
	}

	// Publish will message if WillFlag is set. Server side only.
	if !this.client && this.sess.Cmsg.WillFlag() {
		glog.Infof("(%s) service/stop: connection unexpectedly closed. Sending Will.", this.cid())
		this.onPublish(this.sess.Will)
	}

	// Remove the client topics manager
	if this.client {
		topics.Unregister(this.sess.ID())
	}

	// Remove the session from session store if it's suppose to be clean session
	if this.sess.Cmsg.CleanSession() && this.sessMgr != nil {
		this.sessMgr.Del(this.sess.ID())
	}

	this.conn = nil
	this.in = nil
	this.out = nil
}
