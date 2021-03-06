package mqtt

import (
	"net"
	"sync"
)

//字节池
var bytePool = sync.Pool{
	New:func() interface{} {
		buf := make([]byte, 1)
		return &buf
	},
}

//缓冲层
var bufPool = sync.Pool{
	New:func() interface{} {
		buf := make([]byte,127)
		return &buf
	},
}

// 解析MQTT协议
func GetMqtt(conn net.TCPConn) (*MqttBuffer,*error) {
	mqttBuffer := new(MqttBuffer)
	controlHeader,err := DeCodeControlHeaderFormTCPConn(conn)
	if(err != nil){
		return nil,err
	}
	mqttBuffer.MqttControl = controlHeader
	len,err := DeCodeLenFormTCPConn(conn)
	mqttBuffer.Len = *len
	mqttBuffer.body = DeCodeBodyFormTCPConn(conn,len)
	return mqttBuffer,nil
}

//从TCP链接中获取body
func DeCodeBodyFormTCPConn(conn net.TCPConn,len int)(*[]byte,*error){
	return nil,nil
}

// 从TCP连接中获取一个字节的数据进行解码
func DeCodeLenFormTCPConn(conn net.TCPConn) (*uint,*error) {
	var len uint = 0;
	for i := 1 ; i < 4 ; i++ {
		byteTemp,err := getByte(conn)
		if(err != nil){
			return nil,new(error)
		}
		lenTemp,err := DeCodeLengthByByte((*byteTemp)[0])
		defer bytePool.Put(byteTemp)
		if(err != nil){
			return nil,new(error)
		}
		len |= (lenTemp.Data) << (uint(i * 7))
		if(lenTemp.IsContinue == 0) {
			return &len, nil
		}
	}
	return nil,new(error)
}

// 从TCP连接中获取一个字节的数据进行解码
func DeCodeControlHeaderFormTCPConn(conn net.TCPConn) (*MqttControl,*error) {
	byteTemp,err := getByte(conn)
	if(err != nil){
		return nil,new(error)
	}
	controlHeader,err := DeCodeMqttHeaderByByte((*byteTemp)[0])
	defer bytePool.Put(byteTemp)
	return controlHeader,err
}

// 从网络中读取一个byte的数据
func getByte(conn net.TCPConn) (*[]byte,*error) {
	byteTemp := bytePool.Get().(*[]byte)
	conn.Read(*byteTemp)
	return byteTemp,nil
}

// 从Byte中解析头部
func DeCodeMqttHeaderByByte(buf byte) (*MqttControl,*error) {
	header := new(MqttControl)
	header.Control = (MsgType)(buf & 0xf0)
	header.Dup = buf & 0x08
	header.QoS = (QoS)(buf & 0x06)
	header.Retain = buf & 0x01
	return header,nil
}

// 从byte中解析长度
func DeCodeLengthByByte(buf byte) (*Length,*error) {
	length := new(Length)
	length.IsContinue = (ContinueType)(buf & 0x80)
	length.Data = (uint)(buf & 0x7F)
	return length,nil
}

func (this *MqttControl) GetByte() (*byte,*error)  {
	var buf byte
	buf |= (byte)(this.Control & 0xF0)
	buf |= (byte)(this.Dup & 0x08)
	buf |= (byte)(this.QoS & 0x06)
	buf |= (byte)(this.Retain & 0x01)
	return &buf,nil
}

func GetBufferHeader(buf []byte) (*BufferHeader,*error) {
	bufferHeader := new(BufferHeader)
	controlHeader,_ := DeCodeMqttHeaderByByte(buf[0])
	bufferHeader.MqttControl = * controlHeader
	for i := 0; i < 4 ; i++ {
		length,_ := DeCodeLengthByByte(buf[i+1])
		bufferHeader.Len |= ((length.Data) << (uint)(7 * i))
		if(length.IsContinue == 0){
			bufferHeader.LenIndex = i + 2
			break
		}
	}
	return bufferHeader,nil
}

func GetBytes(length int) ([]byte,*error) {
	var buffers []byte = []byte{}
	for i := 0; i <4 && length > 0 ; i++  {
		buf := (byte)(length & 0x7F);
		length >>= 7
		if(length > 0) {
			buf |= 0x80
		}
		buffers = append(buffers,buf)
	}
	return buffers,nil
}