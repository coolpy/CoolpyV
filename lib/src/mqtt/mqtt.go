package mqtt

func GetHeader(buf uint8) *Header {
	header := new(Header)
	header.Control = (MsgType)(buf & 0xf0);
	header.Dup = buf & 0x08;
	header.QoS = (QoS)(buf & 0x06);
	header.Retain = buf & 0x01;
	return header
}

func GetLength(buf uint8) *Length {
	length := new(Length)
	length.IsContinue = (ContinueType)(buf & 0x80);
	length.Data = buf & 0x7F;
	return length
}

func (this * Header) GetByte() *byte  {
	var buf byte;
	buf |= (byte)(this.Control & 0xF0)
	buf |= (byte)(this.Dup & 0x08)
	buf |= (byte)(this.QoS & 0x06)
	buf |= (byte)(this.Retain & 0x01)
	return &buf
}