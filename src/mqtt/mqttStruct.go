package mqtt

type MqttControl struct {
	Control MsgType
	Dup byte
	QoS QoS
	Retain byte
}

type Length struct  {
	IsContinue ContinueType
	Data uint
}

type MqttBuffer struct {
	*MqttControl
	Len uint
	body []byte
}

type DefaultHeader struct {
	*MqttControl
	*Length
}

type BufferHeader struct {
	MqttControl
	LenIndex int
	Len uint
}