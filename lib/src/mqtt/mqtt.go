package mqtt

func GetControlHeader(buf byte) *ControlHeader {
	header := new(ControlHeader)
	header.Control = (MsgType)(buf & 0xf0)
	header.Dup = buf & 0x08
	header.QoS = (QoS)(buf & 0x06)
	header.Retain = buf & 0x01
	return header
}

func GetLengthByByte(buf byte) *Length {
	length := new(Length)
	length.IsContinue = (ContinueType)(buf & 0x80)
	length.Data = (uint)(buf & 0x7F)
	return length
}

func CheckIsContinue(buf byte) bool{
	return ContinueType(buf & 0x80) == Continue;
}

func (this * ControlHeader) GetByte() *byte  {
	var buf byte
	buf |= (byte)(this.Control & 0xF0)
	buf |= (byte)(this.Dup & 0x08)
	buf |= (byte)(this.QoS & 0x06)
	buf |= (byte)(this.Retain & 0x01)
	return &buf
}

func GetBufferHeader(buf []byte) *BufferHeader {
	bufferHeader := new(BufferHeader)
	bufferHeader.ControlHeader = *GetControlHeader(buf[0])
	for i := 0; i < 4 ; i++ {
		length := GetLengthByByte(buf[i+1])
		bufferHeader.Len |= ((length.Data) << (uint)(7 * i))
		if(length.IsContinue == 0){
			break
		}
	}
	return bufferHeader
}

func GetBytes(length int) []byte{
	var buffers []byte = []byte{}
	for i := 0; i <4 && length > 0 ; i++  {
		buf := (byte)(length & 0x7F);
		length >>= 7
		if(length > 0) {
			buf |= 0x80
		}
		buffers = append(buffers,buf)
	}
	return buffers
}