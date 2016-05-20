package mqtt

import "testing"

func TestPathClean(t *testing.T) {
	header := GetHeader(0x1B)
	if(header.Control != Connect) {
		t.Error("不是连接")
	}
	if(header.Retain != Retain){
		t.Error("不是Retain")
	}
	if(header.QoS != Qos1){
		t.Error("不是Qos1")
	}
	if(header.Dup != Dup){
		t.Error("不是Dup")
	}
}
