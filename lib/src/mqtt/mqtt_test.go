package mqtt

import (
	"testing"
	"fmt"
	"time"
)

const size = 10000000

func TestGetBytesPush(t *testing.T) {
	t0 := time.Now()
	for i := 1; i < size; i++ {
		_ = GetBytes(128)
	}
	t1 := time.Now()
	fmt.Printf("The testIntVectorPush call took %v to run.\n", t1.Sub(t0))
}

func TestGetHeaderPush(t *testing.T) {
	t0 := time.Now()
	for i := 1; i < size; i++ {
		header := GetHeader(0x1B)
		if(header.Control != Connect) {

		}
	}
	t1 := time.Now()
	fmt.Printf("The TestGetHeader call took %v to run.\n", t1.Sub(t0))
}
3
func TestGetHeader(t *testing.T) {
	header := GetHeader(0x1B)
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
	length := GetLength(0x84)
	if(length.IsContinue != Continue) {
		t.Error("不是继续")
	}
	if(length.Data != 4 ){
		t.Error("数据错误")
	}
}

func TestHeader_GetByte(t *testing.T) {
	header := new(Header)
	header.Control = Connect
	header.Dup = Dup
	header.QoS = Qos1
	header.Retain = Retain
	buf := header.GetByte();
	if((*buf) != 0x1B){
		t.Error("转换出错",*buf)
	}
}

func TestGetBytes(t *testing.T) {
	bufs := GetBytes(128)
	if(len(bufs) != 2) {
		t.Error("出错",bufs)
	}
}