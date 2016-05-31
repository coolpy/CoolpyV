package mqtt


type QoS byte
type MsgType byte
type ContinueType byte

const (
	Qos0 QoS = 0x00
	Qos1 QoS = 0x02
	Qos2 QoS = 0x04
)

const (
	Dup byte = 0x08
	Retain byte = 0x01
)

const (
	Connect MsgType = 0x010
	ConNack MsgType = 0x020
	Publish MsgType = 0x030
	PuBack MsgType = 0x040
	PubRec MsgType = 0x050
	PubRel MsgType = 0x060
	PubComp MsgType = 0x070
	SubScribe MsgType = 0x080
	SubAck MsgType = 0x090
	UnSubScribe MsgType = 0x0A0
	UnSubAck MsgType = 0x0B0
	PingReq MsgType = 0x0C0
	PingResp MsgType = 0x0D0
	DisConnect MsgType = 0x0E0
	ExtEnd MsgType = 0x0F0
)

const (
	Continue ContinueType = 0x80
	NotContinue ContinueType = 0x00
)

type ControlHeader struct {
	Control MsgType
	Dup byte
	QoS QoS
	Retain byte
}

type Length struct  {
	IsContinue ContinueType
	Data uint
}

type DefaultHeader struct {
	*ControlHeader
	*Length
}

type BufferHeader struct {
	ControlHeader
	LenIndex int
	Len uint
}