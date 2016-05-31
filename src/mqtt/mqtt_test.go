package mqtt

import (
	"testing"
	"fmt"
	"time"
	"sync"
)

const size = 10000000

func TestGetDefaultHeader(t *testing.T) {
	defaultHeader,_ := GetDefaultHeader([]byte { 0x1B , 0x04})
	if(defaultHeader.Control != Connect){
		t.Error("解析错误 Connect")
	}
	if(defaultHeader.IsContinue != NotContinue){
		t.Error("解析错误 NotContinue")
	}
}

func TestCheckIsContinue(t *testing.T) {
	isContinue,_ := CheckIsContinue(0x80)
	if(!isContinue){
		t.Error("解析失败")
	}
}

func TestGetBytesPush(t *testing.T) {
	t0 := time.Now()
	for i := 1; i < size; i++ {
		_,_ = GetBytes(128)
	}
	t1 := time.Now()
	fmt.Printf("The testIntVectorPush call took %v to run.\n", t1.Sub(t0))
}

func TestIntVectorPush1(t *testing.T) {
	var buf []byte = []byte{ 0x1B }
	bufs,_ := GetBytes(4)
	buf = append(buf,bufs...)
	t0 := time.Now()
	for i := 1; i < size; i++ {
		var buf1 []byte = make([]byte,1024)
		copy(buf1,buf)
		header,_ := GetBufferHeader(buf1)
		if(header.Control != Connect) {

		}
	}
	t1 := time.Now()
	fmt.Printf("The testIntVectorPush1 call took %v to run.\n", t1.Sub(t0))
}

func TestIntVectorPush2(t *testing.T)  {
	p := &sync.Pool{
		New: func() interface{} {
			b := make([]byte,64*1024)
			return &b
		},
	}
	var buf []byte = []byte{ 0x1B }
	bufs,_ := GetBytes(4)
	buf = append(buf,bufs...)
	t0 := time.Now()
	for i := 1; i < size; i++ {
		buf1 := p.Get().(*[]byte)
		copy(*buf1,buf)
		header,_ := GetBufferHeader(*buf1)
		if(header.Control != Connect) {

		}
		p.Put(buf1)
	}
	t1 := time.Now()
	fmt.Printf("The testIntVectorPush2 call took %v to run.\n", t1.Sub(t0))
}

func TestGetHeaderPush(t *testing.T) {
	t0 := time.Now()
	for i := 1; i < size; i++ {
		header,_ := GetControlHeader(0x1B)
		if(header.Control != Connect) {

		}
	}
	t1 := time.Now()
	fmt.Printf("The TestGetHeader call took %v to run.\n", t1.Sub(t0))
}

func TestGetBufferHeader(t *testing.T) {
	var buf []byte = []byte{ 0x1B }
	bufs,_ := GetBytes(123456789)
	buf = append(buf,bufs...)
	bufferHeader,_ := GetBufferHeader(buf)
	if(bufferHeader.Control != Connect) {
		t.Error("不是连接")
	}
	if(bufferHeader.Retain != Retain) {
		t.Error("不是Retain")
	}
	if(bufferHeader.QoS != Qos1) {
		t.Error("不是Qos1")
	}
	if(bufferHeader.Dup != Dup) {
		t.Error("不是Dup")
	}
	if(bufferHeader.Len != 123456789) {
		t.Error(bufferHeader.Len)
	}
}

func TestGetHeader(t *testing.T) {
	header,_ := GetControlHeader(0x1B)
	if(header.Control != Connect) {
		t.Error("不是连接")
	}
	if(header.Retain != Retain) {
		t.Error("不是Retain")
	}
	if(header.QoS != Qos1) {
		t.Error("不是Qos1")
	}
	if(header.Dup != Dup) {
		t.Error("不是Dup")
	}
}

func TestGetLength(t *testing.T) {
	length,_ := GetLengthByByte(0x84)
	if(length.IsContinue != Continue) {
		t.Error("不是继续")
	}
	if(length.Data != 4 ){
		t.Error("数据错误")
	}
}

func TestHeader_GetByte(t *testing.T) {
	header := new(ControlHeader)
	header.Control = Connect
	header.Dup = Dup
	header.QoS = Qos1
	header.Retain = Retain
	buf,_ := header.GetByte();
	if((*buf) != 0x1B){
		t.Error("转换出错",*buf)
	}
}

func TestGetBytes(t *testing.T) {
	bufs,_ := GetBytes(128)
	if(len(bufs) != 2) {
		t.Error("出错",bufs)
	}
}