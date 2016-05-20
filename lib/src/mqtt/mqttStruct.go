package mqtt


type QoS byte
type MsgType byte

const (
	Qos0 QoS = 0x00;
	Qos1 QoS = 0x02;
	Qos2 QoS = 0x04;
)

const (
	Dup byte = 0x08;
	Retain byte = 0x01;
)

const (
	Connect MsgType = 0x01
	ConNack MsgType = 0x02
	Publish MsgType = 0x03
	PuBack MsgType = 0x04
	PubRec MsgType = 0x05
	PubRel MsgType = 0x06
	PubComp MsgType = 0x07
	SubScribe MsgType = 0x08
	SubAck MsgType = 0x09
	UnSubScribe MsgType = 0x0A
	UnSubAck MsgType = 0x0B
	PingReq MsgType = 0x0C
	PingResp MsgType = 0x0D
	DisConnect MsgType = 0x0E
	ExtEnd MsgType = 0x0F
)

type Header struct {
	Control MsgType
	Dup byte
	QoS QoS
	Retain byte
}