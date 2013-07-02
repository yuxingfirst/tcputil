package tcputil

import (
	"errors"
	"reflect"
	"unsafe"
)

//
// 内存池需要实现的接口
//
type MemPool interface {
	Alloc(size int) []byte
}

//
// 简单的内存池实现，用于避免频繁的零散内存申请
//
type SimpleMemPool struct {
	memPool     []byte
	memPoolSize int
	maxPackSize int
}

//
// 创建一个简单内存池，预先申请'memPoolSize'大小的内存，每次分配内存时从中切割出来，直到剩余空间不够分配，再重新申请一块。
// 参数'maxPackSize'用于限制外部申请内存允许的最大长度，所以这个值必须大于等于'memPoolSize'。
//
func NewSimpleMemPool(memPoolSize, maxPackSize int) (*SimpleMemPool, error) {
	if maxPackSize > memPoolSize {
		return nil, errors.New("maxPackSize > memPoolSize")
	}

	return &SimpleMemPool{
		memPool:     make([]byte, memPoolSize),
		memPoolSize: memPoolSize,
		maxPackSize: maxPackSize,
	}, nil
}

//
// 申请一块内存，如果'size'超过'maxPackSize'设置将返回nil
//
func (this *SimpleMemPool) Alloc(size int) (result []byte) {
	if size > this.maxPackSize {
		return nil
	}

	if len(this.memPool) < size {
		this.memPool = make([]byte, this.memPoolSize)
	}

	result = this.memPool[0:size]
	this.memPool = this.memPool[size:]

	return result
}

func getUint(buff []byte, pack int) int {
	var ptr = unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&buff)).Data)

	switch pack {
	case 1:
		return int(buff[0])
	case 2:
		return int(*(*uint16)(ptr))
	case 4:
		return int(*(*uint32)(ptr))
	case 8:
		return int(*(*uint64)(ptr))
	}

	return 0
}

func setUint(buff []byte, pack, value int) {
	var ptr = unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&buff)).Data)

	switch pack {
	case 1:
		buff[0] = byte(value)
	case 2:
		*(*uint16)(ptr) = uint16(value)
	case 4:
		*(*uint32)(ptr) = uint32(value)
	case 8:
		*(*uint64)(ptr) = uint64(value)
	}
}

func getUint16(target []byte) uint16 {
	return *(*uint16)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&target)).Data))
}

func setUint16(target []byte, value uint16) {
	*(*uint16)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&target)).Data)) = value
}

func getUint32(target []byte) uint32 {
	return *(*uint32)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&target)).Data))
}

func setUint32(target []byte, value uint32) {
	*(*uint32)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&target)).Data)) = value
}
