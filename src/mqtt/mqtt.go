package mqtt

func GetControlHeader(buf byte) (*ControlHeader,*error) {
	header := new(ControlHeader)
	header.Control = (MsgType)(buf & 0xf0)
	header.Dup = buf & 0x08
	header.QoS = (QoS)(buf & 0x06)
	header.Retain = buf & 0x01
	return header,nil
}

func GetDefaultHeader(buf []byte) (*DefaultHeader,*error) {
	defaultHeader :=new(DefaultHeader)
	defaultHeader.ControlHeader,_ = GetControlHeader((buf)[0])
	defaultHeader.Length,_ = GetLengthByByte((buf)[1])
	return defaultHeader,nil
}

func GetLengthByByte(buf byte) (*Length,*error) {
	length := new(Length)
	length.IsContinue = (ContinueType)(buf & 0x80)
	length.Data = (uint)(buf & 0x7F)
	return length,nil
}

func CheckIsContinue(buf byte) (bool,*error){
	return ContinueType(buf & 0x80) == Continue,nil;
}

func (this *ControlHeader) GetByte() (*byte,*error)  {
	var buf byte
	buf |= (byte)(this.Control & 0xF0)
	buf |= (byte)(this.Dup & 0x08)
	buf |= (byte)(this.QoS & 0x06)
	buf |= (byte)(this.Retain & 0x01)
	return &buf,nil
}

func GetBufferHeader(buf []byte) (*BufferHeader,*error) {
	bufferHeader := new(BufferHeader)
	controlHeader,_ := GetControlHeader(buf[0])
	bufferHeader.ControlHeader = * controlHeader
	for i := 0; i < 4 ; i++ {
		length,_ := GetLengthByByte(buf[i+1])
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