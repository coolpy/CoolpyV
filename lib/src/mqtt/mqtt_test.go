package mqtt

import "testing"

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