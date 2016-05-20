package mqtt

func GetHeader(buf uint8) *Header {
	header := new(Header)
	header.Control = (MsgType)(buf >> 4);
	header.Dup = buf & 0x08;
	header.QoS = (QoS)(buf & 0x06);
	header.Retain = buf & 0x01;
	return header
}