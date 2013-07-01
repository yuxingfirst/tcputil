package tcp

import (
	"reflect"
	"unsafe"
)

type TcpOutput struct {
	owner *TcpConn
	buff  []byte
	Data  []byte
}

func (this *TcpOutput) Send() error {
	return this.owner.sendRaw(this.buff)
}

func (this *TcpOutput) WriteUint(pack int, value uint64) *TcpOutput {
	switch pack {
	case 1:
		this.WriteUint8(uint8(value))
	case 2:
		this.WriteUint16(uint16(value))
	case 4:
		this.WriteUint32(uint32(value))
	case 8:
		this.WriteUint64(uint64(value))
	default:
		panic("pack != 1 || pack != 2 || pack != 4 || pack != 8")
	}

	return this
}

func (this *TcpOutput) WriteInt8(value int8) *TcpOutput {
	if len(this.Data) < 1 {
		panic("index out of range")
	}
	this.Data[0] = byte(value)
	this.Data = this.Data[1:]
	return this
}

func (this *TcpOutput) WriteUint8(value uint8) *TcpOutput {
	if len(this.Data) < 1 {
		panic("index out of range")
	}
	this.Data[0] = byte(value)
	this.Data = this.Data[1:]
	return this
}

func (this *TcpOutput) WriteInt16(value int16) *TcpOutput {
	if len(this.Data) < 2 {
		panic("index out of range")
	}
	*(*int16)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&this.Data)).Data)) = value
	this.Data = this.Data[2:]
	return this
}

func (this *TcpOutput) WriteUint16(value uint16) *TcpOutput {
	if len(this.Data) < 2 {
		panic("index out of range")
	}
	*(*uint16)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&this.Data)).Data)) = value
	this.Data = this.Data[2:]
	return this
}

func (this *TcpOutput) WriteInt32(value int32) *TcpOutput {
	if len(this.Data) < 4 {
		panic("index out of range")
	}
	*(*int32)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&this.Data)).Data)) = value
	this.Data = this.Data[4:]
	return this
}

func (this *TcpOutput) WriteUint32(value uint32) *TcpOutput {
	if len(this.Data) < 4 {
		panic("index out of range")
	}
	*(*uint32)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&this.Data)).Data)) = value
	this.Data = this.Data[4:]
	return this
}

func (this *TcpOutput) WriteInt64(value int64) *TcpOutput {
	if len(this.Data) < 8 {
		panic("index out of range")
	}
	*(*int64)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&this.Data)).Data)) = value
	this.Data = this.Data[8:]
	return this
}

func (this *TcpOutput) WriteUint64(value uint64) *TcpOutput {
	if len(this.Data) < 8 {
		panic("index out of range")
	}
	*(*uint64)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&this.Data)).Data)) = value
	this.Data = this.Data[8:]
	return this
}

func (this *TcpOutput) WriteBytes(data []byte) *TcpOutput {
	if len(this.Data) < len(data) {
		panic("index out of range")
	}
	copy(this.Data, data)
	this.Data = this.Data[len(data):]
	return this
}

func (this *TcpOutput) WriteBytes8(data []byte) *TcpOutput {
	this.WriteUint8(uint8(len(data)))
	this.WriteBytes(data)
	return this
}

func (this *TcpOutput) WriteBytes16(data []byte) *TcpOutput {
	this.WriteUint16(uint16(len(data)))
	this.WriteBytes(data)
	return this
}

func (this *TcpOutput) WriteBytes32(data []byte) *TcpOutput {
	this.WriteUint32(uint32(len(data)))
	this.WriteBytes(data)
	return this
}

type TcpInput struct {
	Data []byte
}

func NewTcpInput(data []byte) *TcpInput {
	return &TcpInput{data}
}

func (this *TcpInput) Seek(n int) *TcpInput {
	this.Data = this.Data[n:]
	return this
}

func (this *TcpInput) ReadInt8() int8 {
	var result = int8(this.Data[0])
	this.Data = this.Data[1:]
	return result
}

func (this *TcpInput) ReadUint8() uint8 {
	var result = uint8(this.Data[0])
	this.Data = this.Data[1:]
	return result
}

func (this *TcpInput) ReadInt16() int16 {
	if len(this.Data) < 2 {
		panic("index out of range")
	}
	var result = *(*int16)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&this.Data)).Data))
	this.Data = this.Data[2:]
	return result
}

func (this *TcpInput) ReadUint16() uint16 {
	if len(this.Data) < 2 {
		panic("index out of range")
	}
	var result = *(*uint16)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&this.Data)).Data))
	this.Data = this.Data[2:]
	return result
}

func (this *TcpInput) ReadInt32() int32 {
	if len(this.Data) < 4 {
		panic("index out of range")
	}
	var result = *(*int32)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&this.Data)).Data))
	this.Data = this.Data[4:]
	return result
}

func (this *TcpInput) ReadUint32() uint32 {
	if len(this.Data) < 4 {
		panic("index out of range")
	}
	var result = *(*uint32)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&this.Data)).Data))
	this.Data = this.Data[4:]
	return result
}

func (this *TcpInput) ReadInt64() int64 {
	if len(this.Data) < 8 {
		panic("index out of range")
	}
	var result = *(*int64)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&this.Data)).Data))
	this.Data = this.Data[8:]
	return result
}

func (this *TcpInput) ReadUint64() uint64 {
	if len(this.Data) < 8 {
		panic("index out of range")
	}
	var result = *(*uint64)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&this.Data)).Data))
	this.Data = this.Data[8:]
	return result
}

func (this *TcpInput) ReadBytes(n int) []byte {
	var result = this.Data[:n]
	this.Data = this.Data[n:]
	return result
}

func (this *TcpInput) ReadBytes8() []byte {
	var n = this.ReadUint8()
	return this.ReadBytes(int(n))
}

func (this *TcpInput) ReadBytes16() []byte {
	var n = this.ReadUint16()
	return this.ReadBytes(int(n))
}

func (this *TcpInput) ReadBytes32() []byte {
	var n = this.ReadUint32()
	return this.ReadBytes(int(n))
}
